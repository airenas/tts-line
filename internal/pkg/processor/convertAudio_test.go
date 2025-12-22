package processor

import (
	"context"
	"fmt"
	"io"
	"testing"

	"google.golang.org/grpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/airenas/tts-line/internal/pkg/gen/audioconverter"
	"github.com/airenas/tts-line/internal/pkg/service/api"
	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/test/mocks"
)

func TestNewConverter(t *testing.T) {
	pr, err := NewConverter("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestNewConverter_Fails(t *testing.T) {
	pr, err := NewConverter("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeConvert(t *testing.T) {
	mockClient := new(mockAudioConverterClient)
	mockStream := new(mockConvertStreamClient)

	mockClient.On("ConvertStream", mock.Anything, mock.Anything).Return(mockStream, nil)
	mockStream.On("Send", mock.Anything).Return(nil)
	mockStream.On("CloseSend").Return(nil)
	mockStream.On("Recv").Return(&audioconverter.StreamFileReply{Chunk: []byte("mp3")}, nil).Once()
	mockStream.On("Recv").Return(nil, io.EOF)

	pr, _ := NewConverter("http://server")
	assert.NotNil(t, pr)
	pr.(*audioConverter).client = mockClient
	d := synthesizer.TTSData{}
	d.Audio = testGenerateSampleData(t, []byte("wav"))
	d.Input = &api.TTSRequestConfig{OutputMetadata: []string{"olia"}, OutputFormat: api.AudioMP3}
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.Equal(t, []byte("mp3"), d.AudioMP3)
	mockClient.AssertNumberOfCalls(t, "ConvertStream", 1)
	// mockStream.AssertNumberOfCalls(t, "Send", 2)
	mockStream.AssertNumberOfCalls(t, "Recv", 2)
	cData := mockStream.Calls[0].Arguments[0]
	cInp, _ := cData.(*audioconverter.StreamConvertInput)
	require.NotNil(t, cInp.GetMetadata())
	assert.Equal(t, "MP3", cInp.GetMetadata().Format.String())
	assert.Equal(t, []string{"olia"}, cInp.GetMetadata().Metadata)
}

func TestInvokeConvert_Skip(t *testing.T) {
	mockClient := new(mockAudioConverterClient)
	mockStream := new(mockConvertStreamClient)

	mockClient.On("ConvertStream", mock.Anything, mock.Anything).Return(mockStream, nil)

	pr, _ := NewConverter("http://server")
	assert.NotNil(t, pr)
	pr.(*audioConverter).client = mockClient
	d := synthesizer.TTSData{}
	d.Audio = testGenerateSampleData(t, []byte("wav"))
	d.Input = &api.TTSRequestConfig{OutputMetadata: []string{"olia"}, OutputFormat: api.AudioNone}
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)

	mockClient.AssertNumberOfCalls(t, "ConvertStream", 0)
}

func TestInvokeConvert_SkipWAV(t *testing.T) {
	mockClient := new(mockAudioConverterClient)
	mockStream := new(mockConvertStreamClient)

	mockClient.On("ConvertStream", mock.Anything, mock.Anything).Return(mockStream, nil)

	pr, _ := NewConverter("http://server")
	assert.NotNil(t, pr)
	pr.(*audioConverter).client = mockClient
	d := synthesizer.TTSData{}
	d.Audio = testGenerateSampleData(t, []byte("wav"))
	d.Input = &api.TTSRequestConfig{OutputMetadata: []string{"olia"}, OutputFormat: api.AudioWAV}
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)

	mockClient.AssertNumberOfCalls(t, "ConvertStream", 0)
}

func TestInvokeConvert_Fail(t *testing.T) {
	mockClient := new(mockAudioConverterClient)

	mockClient.On("ConvertStream", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))

	pr, _ := NewConverter("http://server")
	assert.NotNil(t, pr)
	pr.(*audioConverter).client = mockClient
	d := synthesizer.TTSData{}
	d.Audio = testGenerateSampleData(t, []byte("wav"))
	d.Input = &api.TTSRequestConfig{OutputMetadata: []string{"olia"}, OutputFormat: api.AudioMP3}
	err := pr.Process(context.TODO(), &d)
	assert.NotNil(t, err)
	assert.Nil(t, d.AudioMP3)
}

func TestInvokeConvert_FailSend(t *testing.T) {
	mockClient := new(mockAudioConverterClient)
	mockStream := new(mockConvertStreamClient)

	mockClient.On("ConvertStream", mock.Anything, mock.Anything).Return(mockStream, nil)
	mockStream.On("Send", mock.Anything).Return(fmt.Errorf("error"))
	mockStream.On("CloseSend").Return(nil)
	mockStream.On("Recv").Return(&audioconverter.StreamFileReply{Chunk: []byte("mp3")}, nil).Times(1000)
	mockStream.On("Recv").Return(nil, io.EOF)

	pr, _ := NewConverter("http://server")
	assert.NotNil(t, pr)
	pr.(*audioConverter).client = mockClient
	d := synthesizer.TTSData{}
	d.Audio = testGenerateSampleData(t, []byte("wav"))
	d.Input = &api.TTSRequestConfig{OutputMetadata: []string{"olia"}, OutputFormat: api.AudioMP3}
	err := pr.Process(context.TODO(), &d)
	assert.NotNil(t, err)
	assert.Nil(t, d.AudioMP3)
}

func TestInvokeConvert_FailSendEOF(t *testing.T) {
	mockClient := new(mockAudioConverterClient)
	mockStream := new(mockConvertStreamClient)

	mockClient.On("ConvertStream", mock.Anything, mock.Anything).Return(mockStream, nil)
	mockStream.On("Send", mock.Anything).Return(nil)
	mockStream.On("CloseSend").Return(fmt.Errorf("error"))
	mockStream.On("Recv").Return(&audioconverter.StreamFileReply{Chunk: []byte("mp3")}, nil).Times(1000)
	mockStream.On("Recv").Return(nil, io.EOF)

	pr, _ := NewConverter("http://server")
	assert.NotNil(t, pr)
	pr.(*audioConverter).client = mockClient
	d := synthesizer.TTSData{}
	d.Audio = testGenerateSampleData(t, []byte("wav"))
	d.Input = &api.TTSRequestConfig{OutputMetadata: []string{"olia"}, OutputFormat: api.AudioMP3}
	err := pr.Process(context.TODO(), &d)
	assert.NotNil(t, err)
	assert.Nil(t, d.AudioMP3)
}

func TestInvokeConvert_FailReceive(t *testing.T) {
	mockClient := new(mockAudioConverterClient)
	mockStream := new(mockConvertStreamClient)

	mockClient.On("ConvertStream", mock.Anything, mock.Anything).Return(mockStream, nil)
	mockStream.On("Send", mock.Anything).Return(nil)
	mockStream.On("CloseSend").Return(nil)
	mockStream.On("Recv").Return(nil, fmt.Errorf("error")).Once()
	mockStream.On("Recv").Return(nil, io.EOF)

	pr, _ := NewConverter("http://server")
	assert.NotNil(t, pr)
	pr.(*audioConverter).client = mockClient
	d := synthesizer.TTSData{}
	d.Audio = testGenerateSampleData(t, []byte("wav"))
	d.Input = &api.TTSRequestConfig{OutputMetadata: []string{"olia"}, OutputFormat: api.AudioMP3}
	err := pr.Process(context.TODO(), &d)
	assert.NotNil(t, err)
	assert.Nil(t, d.AudioMP3)
}

type mockAudioConverterClient struct {
	mock.Mock
}

func (m *mockAudioConverterClient) Convert(ctx context.Context, inp *audioconverter.ConvertInput, opts ...grpc.CallOption) (*audioconverter.ConvertReply, error) {
	args := m.Called(ctx, inp, opts)
	return args.Get(0).(*audioconverter.ConvertReply), args.Error(1)
}

func (m *mockAudioConverterClient) ConvertStream(ctx context.Context, opts ...grpc.CallOption) (audioconverter.AudioConverter_ConvertStreamClient, error) {
	args := m.Called(ctx, opts)
	return mocks.To[audioconverter.AudioConverter_ConvertStreamClient](args.Get(0)), args.Error(1)
}

type mockConvertStreamClient struct {
	mock.Mock
	grpc.ClientStream
}

func (m *mockConvertStreamClient) Send(req *audioconverter.StreamConvertInput) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *mockConvertStreamClient) CloseSend() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockConvertStreamClient) Recv() (*audioconverter.StreamFileReply, error) {
	args := m.Called()
	return mocks.To[*audioconverter.StreamFileReply](args.Get(0)), args.Error(1)
}
