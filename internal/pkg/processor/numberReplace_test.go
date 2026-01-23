package processor

import (
	"context"
	"reflect"
	"testing"

	"github.com/airenas/tts-line/internal/pkg/synthesizer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	httpInvokerMock *mockHTTPInvoker
)

func initTest(t *testing.T) {
	httpInvokerMock = &mockHTTPInvoker{}
}

func TestCreateNumberReplace(t *testing.T) {
	initTest(t)
	pr, err := NewNumberReplace("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestCreateNumberReplace_Fails(t *testing.T) {
	initTest(t)
	pr, err := NewNumberReplace("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeNumberReplace(t *testing.T) {
	initTest(t)
	pr, _ := NewNumberReplace("http://server")
	assert.NotNil(t, pr)
	pr.(*numberReplace).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{NormalizedText: []string{"3"}}
	httpInvokerMock.On("InvokeText", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*string) = "trys"
		}).Return(nil)
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.Equal(t, []string{"trys"}, d.TextWithNumbers)
}

func TestInvokeNumberReplace_Fail(t *testing.T) {
	initTest(t)
	pr, _ := NewNumberReplace("http://server")
	assert.NotNil(t, pr)
	pr.(*numberReplace).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{}
	httpInvokerMock.On("InvokeText", mock.Anything, mock.Anything).Return(errors.New("haha"))
	err := pr.Process(context.TODO(), &d)
	assert.NotNil(t, err)
}

func TestInvokeNumberReplace_Skip(t *testing.T) {
	d := &synthesizer.TTSData{}
	d.Cfg.JustAM = true
	pr, _ := NewNumberReplace("http://server")
	err := pr.Process(context.TODO(), d)
	assert.Nil(t, err)
}

func TestCreateSSMLNumberReplace(t *testing.T) {
	initTest(t)
	pr, err := NewSSMLNumberReplace("http://server")
	assert.Nil(t, err)
	assert.NotNil(t, pr)
}

func TestCreateSSMLNumberReplace_Fails(t *testing.T) {
	initTest(t)
	pr, err := NewSSMLNumberReplace("")
	assert.NotNil(t, err)
	assert.Nil(t, pr)
}

func TestInvokeSSMLNumberReplace(t *testing.T) {
	initTest(t)
	pr, _ := NewSSMLNumberReplace("http://server")
	assert.NotNil(t, pr)
	pr.(*ssmlNumberReplace).httpWrap = httpInvokerMock
	d := synthesizer.TTSData{CleanedText: []string{"3 oli{a/} 4"}}
	httpInvokerMock.On("InvokeText", mock.Anything, mock.Anything).Run(
		func(params mock.Arguments) {
			*params[1].(*string) = "trys olia keturi"
		}).Return(nil)
	err := pr.Process(context.TODO(), &d)
	assert.Nil(t, err)
	assert.Equal(t, []string{"trys oli{a/} keturi"}, d.TextWithNumbers)
}

func TestInvokeSSMLNumberReplace_Real3(t *testing.T) {
	tests := []struct {
		name       string
		in, out    string
		wantErr    bool
		wantString string
	}{
		{name: "real3-1", in: "(1988 m. surinkęs 25 proc. balsų), bjūk{e~}nenas (1996 m. surinkęs 23 proc. balsų) ir f{o/}rbzas (2000 m. surinkęs 31 proc. balsų)",
			out: "(tūkstantis devyni šimtai aštuoniasdešimt aštuntieji metai surinkęs dvidešimt penki procentai balsų), bjūkenenas (tūkstantis devyni šimtai " +
				"devyniasdešimt šeštieji metai surinkęs dvidešimt trys procentai balsų) ir forbzas (du tūkstantieji metai surinkęs trisdešimt vienas procentas balsų)", wantErr: false,
			wantString: "(tūkstantis devyni šimtai aštuoniasdešimt aštuntieji metai surinkęs dvidešimt penki procentai balsų), bjūk{e~}nenas (tūkstantis devyni šimtai devyniasdešimt " +
				"šeštieji metai surinkęs dvidešimt trys procentai balsų) ir f{o/}rbzas (du tūkstantieji metai surinkęs trisdešimt vienas procentas balsų)"},
		{name: "real", in: "kadenciją 1998-2002 m. {o\\}rb buvo demokratiškas vadovas. Kad 2010 m. grįžęs į valdžią jis staiga virto autokratu, buvo didelis",
			out: "kadenciją tūkstantis devyni šimtai devyniasdešimt aštuntaisiais-du tūkstančiai antraisiais metais orb buvo demokratiškas vadovas." +
				" Kad du tūkstančiai dešimtaisiais metais grįžęs į valdžią jis staiga virto autokratu, buvo didelis", wantErr: false,
			wantString: "kadenciją tūkstantis devyni šimtai devyniasdešimt aštuntaisiais-du tūkstančiai antraisiais metais {o\\}rb buvo demokratiškas vadovas." +
				" Kad du tūkstančiai dešimtaisiais metais grįžęs į valdžią jis staiga virto autokratu, buvo didelis"},
		{name: "real1", in: "kadenciją 1998-2002 m. {o\\}rb buvo demokratiškas vadovas. Kad 2010 m. grįžęs į valdžią jis staiga virto autokratu, buvo",
			out: "kadenciją tūkstantis devyni šimtai devyniasdešimt aštuntaisiais-du tūkstančiai antraisiais metais orb buvo demokratiškas vadovas." +
				" Kad du tūkstančiai dešimtaisiais metais grįžęs į valdžią jis staiga virto autokratu, buvo", wantErr: false,
			wantString: "kadenciją tūkstantis devyni šimtai devyniasdešimt aštuntaisiais-du tūkstančiai antraisiais metais {o\\}rb buvo demokratiškas vadovas." +
				" Kad du tūkstančiai dešimtaisiais metais grįžęs į valdžią jis staiga virto autokratu, buvo"},
		{name: "real2", in: "Be to, per šį laikotarpį rusai neteko 3 795 tankų, 7 432 šarvuotųjų kovos mašinų, 3 359 artilerijos sistemų, " +
			"570 reaktyvinės salvinės ugnies sistemų, 327 oro gynybos priemonių, 309 lėktuvų, 296 sraigtasparnių, 2 977 dronų, 1 015 sparnuotųjų raketų, 18 laivų, 6 148 automobilių",
			out: "Be to, per šį laikotarpį rusai neteko trijų tūkstančių septynių šimtų devyniasdešimt penkių tankų, " +
				"septynių tūkstančių keturių šimtų trisdešimt dviejų šarvuotųjų kovos mašinų, trys tūkstančiai trys šimtai penkiasdešimt devintosios artilerijos sistemų, " +
				"penki šimtai septyniasdešimtosios reaktyvinės salvinės ugnies sistemų, trys šimtai dvidešimt septintosios oro gynybos priemonių, " +
				"trijų šimtų devynių lėktuvų, dviejų šimtų devyniasdešimt šešių sraigtasparnių, dviejų tūkstančių devynių šimtų septyniasdešimt septynių dronų, " +
				"tūkstantį penkiolika sparnuotųjų raketų, aštuoniolika laivų, šešių tūkstančių šimto keturiasdešimt aštuonių automobilių", wantErr: false,
			wantString: ""},
		{name: "real3", in: "Robertsonas (1988 m. surinkęs 25 proc. balsų), bjūk{e~}nenas (1996 m. surinkęs 23 proc. balsų) ir f{o/}rbzas (2000 m. surinkęs 31 proc. balsų)",
			out: "Robertsonas (tūkstantis devyni šimtai aštuoniasdešimt aštuntieji metai surinkęs dvidešimt penki procentai balsų), bjūkenenas (tūkstantis devyni šimtai " +
				"devyniasdešimt šeštieji metai surinkęs dvidešimt trys procentai balsų) ir forbzas (du tūkstantieji metai surinkęs trisdešimt vienas procentas balsų)", wantErr: false,
			wantString: "Robertsonas (tūkstantis devyni šimtai aštuoniasdešimt aštuntieji metai surinkęs dvidešimt penki procentai balsų), bjūk{e~}nenas (tūkstantis devyni šimtai devyniasdešimt " +
				"šeštieji metai surinkęs dvidešimt trys procentai balsų) ir f{o/}rbzas (du tūkstantieji metai surinkęs trisdešimt vienas procentas balsų)"},
		{name: "real4", in: `Norvegija n750,00 Eur n n nII grupė nAustrija, Belgija, Vokietija, Prancūzija, Italija, 
		Graikija, Ispanija, Kipras, Nyderlandai, Malta, Portugalija n750,00 Eur n n nIII grupė nBulgarija, Kroatija, Čekija, Estija, 
		Latvija, Vengrija, Lenkija, Rumunija, Serbija, Slovakija, Slovėnija, Šiaurės Makedonija, Turkija n690,00 Eur n n nIV grupė nŠveicarija*
		 n700,00 Eur n n n nŠalims už ES/EEE ribų: nStipendija 700,00 Eur/mėn. nIšmoka kelionei, kurios dydis nustatomas pagal atstumą* nuo Lietuvos 
		 (Kauno miesto) iki praktikos organizacijos vietos (miesto): n n100-499 km - 180,00 EUR n500-1999 km - 275,00 EUR n2000-2999 km - 360,00 EUR 
		 n3000-3999 km - 530,00 EUR n4000-7999 km - 820,00 EUR nvirš 8000 km - 1500,00 EUR n nPapildoma ekologiškos kelionės stipendija, jei keliaujama 
		 autobusu ar traukiniu. nAtrankos rezultatai bus skelbiami el. paštu, informuojant visus kandidatus. nŠis konkursas yra išskirtinė galimybė praktikai 
		 išvykti už ES/EEE ribų su „Erasmus+“ stipendija, kitas toks konkursas bus skelbiamas tik rudens semestro metu. nDaugiau informacijos: n n n`, wantErr: false,
			out: `Norvegija n septyni šimtai penkiasdešimt ,nulis nulis eurų n n nII grupė 
		 nAustrija, Belgija, Vokietija, Prancūzija, Italija, Graikija, Ispanija, Kipras, Nyderlandai, Malta, Portugalija n septyni šimtai penkiasdešimt 
		 ,nulis nulis eurų n n nIII grupė nBulgarija, Kroatija, Čekija, Estija, Latvija, Vengrija, Lenkija, Rumunija, Serbija, Slovakija, Slovėnija, Šiaurės Makedonija, 
		 Turkija n šeši šimtai devyniasdešimt ,nulis nulis eurų n n nIV grupė nŠveicarija* n septyni šimtai ,nulis nulis eurų n n n nŠalims už ES/EEE ribų: nStipendija 
		 septyni šimtai kablelis nulis nulis eurų/mėn. nIšmoka kelionei, kurios dydis nustatomas pagal atstumą* nuo Lietuvos (Kauno miesto) iki praktikos organizacijos 
		 vietos (miesto): n n šimtas -keturi šimtai devyniasdešimt devyni kilometrai - šimtas aštuoniasdešimt kablelis nulis nulis eurų n penki šimtai -tūkstantis devyni 
		 šimtai devyniasdešimt devyni kilometrai - du šimtai septyniasdešimt penki kablelis nulis nulis eurai n du tūkstančiai -du tūkstančiai devyni šimtai devyniasdešimt 
		 devyni kilometrai - trys šimtai šešiasdešimt kablelis nulis nulis eurų n trys tūkstančiai -trys tūkstančiai devyni šimtai devyniasdešimt devyni kilometrai - penki 
		 šimtai trisdešimt kablelis nulis nulis eurų n keturi tūkstančiai -septyni tūkstančiai devyni šimtai devyniasdešimt devyni kilometrai - aštuoni šimtai dvidešimt kablelis 
		 nulis nulis eurų nvirš aštuoni tūkstančiai kilometrų - tūkstantis penki šimtai kablelis nulis nulis eurų n nPapildoma ekologiškos kelionės stipendija, jei keliaujama autobusu 
		 ar traukiniu. nAtrankos rezultatai bus skelbiami el. paštu, informuojant visus kandidatus. nŠis konkursas yra išskirtinė galimybė praktikai išvykti už ES/EEE ribų su „Erasmus+“ 
		 stipendija, kitas toks konkursas bus skelbiamas tik rudens semestro metu. nDaugiau informacijos: n n n`, wantString: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if tt.name != "real3" {
			// 	return
			// }
			initTest(t)
			pr, _ := NewSSMLNumberReplace("http://server")
			assert.NotNil(t, pr, "err on "+tt.name)
			pr.(*ssmlNumberReplace).httpWrap = httpInvokerMock
			d := synthesizer.TTSData{CleanedText: []string{tt.in}}
			httpInvokerMock.On("InvokeText", mock.Anything, mock.Anything).Run(
				func(params mock.Arguments) {
					*params[1].(*string) = tt.out
				}).Return(nil)
			err := pr.Process(context.TODO(), &d)
			assert.Equal(t, tt.wantErr, err != nil, "err on "+tt.name+":%s", err)
			if tt.wantString != "" {
				assert.Equal(t, []string{tt.wantString}, d.TextWithNumbers, "err on "+tt.name)
			}
		})
	}
}

func Test_mapAccentsBack(t *testing.T) {
	type args struct {
		new  string
		orig []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{name: "several lines", args: args{new: "a b c d e", orig: []string{"a b", "c d"}}, want: []string{"a b", "c d e"}, wantErr: false},
		{name: "drop last", args: args{new: "a b c", orig: []string{"a b c d"}}, want: []string{"a b c"}, wantErr: false},
		{name: "add last", args: args{new: "a b c d e f", orig: []string{"a b c"}}, want: []string{"a b c d e f"}, wantErr: false},
		{name: "empty", args: args{new: "", orig: []string{""}}, want: []string{""}, wantErr: false},
		{name: "no accent", args: args{new: "a b c d r", orig: []string{"a b c d r"}}, want: []string{"a b c d r"}, wantErr: false},
		{name: "with accent", args: args{new: "a b c d r", orig: []string{"a {b/} c d {r~}"}}, want: []string{"a {b/} c d {r~}"}, wantErr: false},
		{name: "several lines2", args: args{new: "a b c d e f g h", orig: []string{"a b", "c d", "g h"}}, want: []string{"a b", "c d e f", "g h"}, wantErr: false},
		{name: "fail on other word", args: args{new: "a b c d k", orig: []string{"a {b/} c d {r~}"}}, want: nil, wantErr: true},
		{name: "fail on missing", args: args{new: "a b c d", orig: []string{"a {b/} c d {r~}"}}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapAccentsBack(t.Context(), tt.args.new, tt.args.orig)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapAccentsBack() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapAccentsBack() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockHTTPInvoker struct{ mock.Mock }

func (m *mockHTTPInvoker) InvokeText(ctx context.Context, in string, out interface{}) error {
	args := m.Called(in, out)
	return args.Error(0)
}
