package processor

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/airenas/tts-line/internal/pkg/gen/audioconverter"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	grpcCodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type audioConverter struct {
	client   audioconverter.AudioConverterClient
	timeout  time.Duration
	backoffF func() backoff.BackOff
	retryF   func(error) bool
}

// NewConverter creates new processor for wav to mp3/m4a conversion
func NewConverter(urlStr string) (synthesizer.Processor, error) {
	if urlStr == "" {
		return nil, fmt.Errorf("empty url")
	}
	res := &audioConverter{
		timeout: time.Second * 120,
	}
	conn, err := grpc.NewClient(urlStr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to gRPC server: %w", err)
	}
	res.client = audioconverter.NewAudioConverterClient(conn)
	res.backoffF = newSimpleBackoff
	res.retryF = isRetryableGRPCError
	return res, nil
}

func (p *audioConverter) Process(ctx context.Context, data *synthesizer.TTSData) error {
	ctx, span := utils.StartSpan(ctx, "audioConverter.Process")
	defer span.End()

	if data.Input.OutputFormat == api.AudioNone {
		return nil
	}
	if data.Input.OutputFormat == api.AudioWAV {
		log.Ctx(ctx).Debug().Msg("No conversion needed to WAV")
		data.AudioMP3 = data.Audio
		return nil
	}

	audio, err := p.invoke(ctx, data)
	if err != nil {
		return fmt.Errorf("convert audio: %w", err)
	}
	log.Ctx(ctx).Debug().Int("len", len(audio)).Msg("Audio conversion done")
	data.AudioMP3 = audio
	return nil
}

func (p *audioConverter) invoke(ctx context.Context, data *synthesizer.TTSData) ([]byte, error) {
	af, err := makeAudioConverterFormat((data.Input.OutputFormat))
	if err != nil {
		return nil, fmt.Errorf("convert audio format: %w", err)
	}

	ctx, cf := context.WithTimeout(ctx, p.timeout)
	defer cf()

	ctx, span := utils.StartSpan(ctx, "audioConverter.invoke")
	defer span.End()

	failC := 0
	var res []byte
	op := func() error {
		res, err = p.invokeInternal(ctx, data, af, failC)
		if err != nil {
			failC++
			if !p.retryF(err) {
				return backoff.Permanent(err)
			}
			select {
			case <-ctx.Done(): // do not retry if context is done
				errCtx := ctx.Err()
				if errCtx != nil && err != errCtx {
					err = fmt.Errorf("%w: %w", errCtx, err)
				}
				return backoff.Permanent(err)
			default:
			}
			log.Ctx(ctx).Warn().Int("retry", failC).Err(err).Msg("failed to convert audio")
			return err
		}
		return nil
	}
	err = backoff.Retry(op, p.backoffF())
	if err == nil && failC > 0 {
		log.Ctx(ctx).Info().Int("times", failC).Msg("Success after retrying")
	}
	return res, err
}

func (p *audioConverter) invokeInternal(ctx context.Context, data *synthesizer.TTSData, audioFormat audioconverter.AudioFormat, retry int) (res []byte, err error) {
	if retry > 0 {
		log.Ctx(ctx).Debug().Int("retry", retry).Msg("Retrying audio conversion")
	}
	ctx, span := utils.StartSpan(ctx, "audioConverter.invokeInternal", trace.WithSpanKind(trace.SpanKindClient))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		span.End()
	}()

	stream, err := p.client.ConvertStream(utils.InjectTraceToGRPC(ctx))
	if err != nil {
		return nil, fmt.Errorf("create stream: %w", err)
	}

	err = stream.Send(&audioconverter.StreamConvertInput{
		Payload: &audioconverter.StreamConvertInput_Metadata{
			Metadata: &audioconverter.InitialMetadata{
				Format:   audioFormat,
				Metadata: data.Input.OutputMetadata,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("send metadata: %w", err)
	}

	const chunkSize = 256 * 1024 // 256KB

	errChan := make(chan error, 1)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		audioData := data.Audio
		for start := 0; start < len(audioData); start += chunkSize {
			end := start + chunkSize
			if end > len(audioData) {
				end = len(audioData)
			}

			err := stream.Send(&audioconverter.StreamConvertInput{
				Payload: &audioconverter.StreamConvertInput_Chunk{
					Chunk: audioData[start:end],
				},
			})
			if err != nil {
				errChan <- fmt.Errorf("send audio chunk: %w", err)
				return
			}
		}

		err := stream.CloseSend()
		if err != nil {
			errChan <- fmt.Errorf("close send stream: %w", err)
			return
		}
	}()

	log.Ctx(ctx).Trace().Msg("receiving audio")

read:
	for {
		select {
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled: %w", ctx.Err())
		default:
			reply, err := stream.Recv()
			if err == io.EOF {
				break read
			}
			if err != nil {
				return nil, fmt.Errorf("failed to receive stream reply: %w", err)
			}
			res = append(res, reply.Chunk...)
		}
	}
	wg.Wait()
	log.Ctx(ctx).Trace().Msg("received audio")
	return res, nil
}

func makeAudioConverterFormat(audioFormatEnum api.AudioFormatEnum) (audioconverter.AudioFormat, error) {
	if audioFormatEnum == api.AudioMP3 {
		return audioconverter.AudioFormat_MP3, nil
	}
	if audioFormatEnum == api.AudioM4A {
		return audioconverter.AudioFormat_M4A, nil
	}
	if audioFormatEnum == api.AudioULAW {
		return audioconverter.AudioFormat_ULAW, nil
	}
	return audioconverter.AudioFormat_AUDIO_FORMAT_UNSPECIFIED, fmt.Errorf("unknown audio format: %s", audioFormatEnum)
}

// Info return info about processor
func (p *audioConverter) Info() string {
	return "audioConverter"
}

func isRetryableGRPCError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error — maybe retry (optional)
		return false
	}
	switch st.Code() {
	case grpcCodes.Unavailable, grpcCodes.DeadlineExceeded, grpcCodes.ResourceExhausted,
		grpcCodes.Aborted, grpcCodes.Internal, grpcCodes.Unknown:
		return true
	default:
		return false
	}
}
