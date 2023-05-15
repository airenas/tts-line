package processor

import (
	"fmt"
	"strings"
	"time"

	"github.com/airenas/go-app/pkg/goapp"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/airenas/tts-line/internal/pkg/utils"
)

type normalizer struct {
	httpWrap HTTPInvokerJSON
}

// NewNormalizer creates new text normalize processor
func NewNormalizer(urlStr string) (synthesizer.Processor, error) {
	res := &normalizer{}
	var err error
	res.httpWrap, err = newHTTPWrapBackoff(urlStr, time.Second*10)
	if err != nil {
		return nil, fmt.Errorf("can't init http client: %w", err)
	}
	return res, nil
}

func (p *normalizer) Process(data *synthesizer.TTSData) error {
	if p.skip(data) {
		goapp.Log.Info("Skip normalize")
		return nil
	}
	defer goapp.Estimate("Normalize")()
	txt := strings.Join(data.CleanedText, "")
	utils.LogData("Input: ", txt)
	inData := &normRequestData{Orig: txt}
	var output normResponseData
	err := p.httpWrap.InvokeJSON(inData, &output)
	if err != nil {
		return fmt.Errorf("normalize (%s): %w", output.Err, err)
	}

	data.NormalizedText, err = processNormalizedOutput(output, data.CleanedText)
	if err != nil {
		return err
	}
	utils.LogData("Output: ", strings.Join(data.NormalizedText, ""))
	return nil
}

func processNormalizedOutput(output normResponseData, input []string) ([]string, error) {
	if len(input) == 1 {
		return []string{output.Res}, nil
	}
	if len(output.Rep) == 0 {
		return input, nil
	}
	resR := make([][]rune, len(input))
	for i := range input {
		resR[i] = []rune(input[i])
	}
	
	shift, fromI := 0, 0
	for _, rep := range output.Rep {
		var atI int
		var err error


		atI, fromI, err = getNextStr(resR, fromI, rep.Beg - shift)
		if (err != nil) {
			return nil, fmt.Errorf("err replace %v: %w", rep, err)
		}
		rns := resR[fromI]
		repRns := []rune(rep.Text)
		rnsNew := append(rns[:atI], repRns...)
		end := (rep.End - rep.Beg) + atI
		if end < len(rns) {
			rnsNew = append(rnsNew, rns[end:]...)
		} 
		resR[fromI] = rnsNew
		rem := end - len(rns)
		for rem > 0 {
			fromI++
			if len(resR[fromI]) > rem {
				resR[fromI] = resR[fromI][rem:]
				rem = 0
			} else {
				rem -= len(resR[fromI])
				resR[fromI] = nil
			}
		}
		shift += (rep.End - rep.Beg) - len(repRns)
	}
	
	res := make ([]string, len(resR))
	for i := range input {
		res[i] = string(resR[i])
	}
	return res, nil
}

func getNextStr(resR [][]rune, fromI, at int) (int, int, error){
	for i := fromI; i < len(resR); i++ {
		res := resR[i]
		l := len(res)
		if at > l {
			fromI++
			at -= l
		} else {
			return at, fromI, nil
		}
	}
	return 0, 0, fmt.Errorf("wrong start pos at [%d]%d", fromI, at)
}

type normRequestData struct {
	Orig string `json:"org"`
}

type normResponseData struct {
	Err string                `json:"err"`
	Org string                `json:"org"`
	Rep []normResponseDataRep `json:"rep"`
	Res string                `json:"res"`
}

type normResponseDataRep struct {
	Beg  int  `json:"beg"`
	End  int  `json:"end"`
	Text string `json:"text"`
}

func (p *normalizer) skip(data *synthesizer.TTSData) bool {
	return data.Cfg.JustAM
}

// Info return info about processor
func (p *normalizer) Info() string {
	return fmt.Sprintf("normalizer(%s)", utils.RetrieveInfo(p.httpWrap))
}
