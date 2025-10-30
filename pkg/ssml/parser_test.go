package ssml

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	vf := func(s string) (string, error) {
		return s, nil
	}
	tests := []struct {
		name    string
		xml     string
		vf      func(string) (string, error)
		want    []Part
		wantErr bool
	}{
		{name: "simple empty", xml: "<speak></speak>", want: []Part{}, wantErr: false},
		{name: "simple", xml: "<speak>olia</speak>", want: []Part{
			&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "olia"}}}},
			wantErr: false},
		{name: "speak fail", xml: "aa<speak>olia</speak>", want: []Part{},
			wantErr: true},
		{name: "speak fail", xml: "<speak>olia</speak><speak>", want: []Part{},
			wantErr: true},
		{name: "speak fail", xml: "<speak>olia<speak>", want: []Part{},
			wantErr: true},
		{name: "speak fail", xml: "<speak>olia</speak>olia", want: []Part{},
			wantErr: true},
		{name: "speak fail", xml: "<b>olia</b>", want: []Part{},
			wantErr: true},
		{name: "empty", wantErr: true},
		{name: "<p>", xml: "<speak><p>olia</p></speak>", want: []Part{
			&Pause{Duration: pDuration},
			&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "olia"}}},
			&Pause{Duration: pDuration}},
			wantErr: false},
		{name: "<p>", xml: "<speak><p>olia</p></speak>", want: []Part{
			&Pause{Duration: pDuration},
			&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "olia"}}},
			&Pause{Duration: pDuration}},
			wantErr: false},
		{name: "<break> time", xml: `<speak><break time="10s"/></speak>`, want: []Part{
			&Pause{Duration: time.Second * 10, IsBreak: true}},
			wantErr: false},
		{name: "<break> strength", xml: `<speak><break strength="x-weak"/></speak>`, want: []Part{
			&Pause{Duration: time.Millisecond * 250, IsBreak: true},
		}, wantErr: false},
		{name: "<break> with text", xml: `<speak><break strength="x-weak">olia</break></speak>`,
			want: []Part{}, wantErr: true},
		{name: "<voice> strength", xml: `<speak><voice name="ooo">aaa</voice></speak>`,
			want: []Part{
				&Text{Voice: "ooo", Speed: 1, Texts: []TextPart{{Text: "aaa"}}},
			}, wantErr: false},
		{name: "<voice> inner", xml: `<speak>
		<voice name="ooo">aaa
		<voice name="ooo1">aaa1</voice>
		end
		</voice>
		end def</speak>`,
			want: []Part{
				&Text{Voice: "ooo", Speed: 1, Texts: []TextPart{{Text: "aaa"}}},
				&Text{Voice: "ooo1", Speed: 1, Texts: []TextPart{{Text: "aaa1"}}},
				&Text{Voice: "ooo", Speed: 1, Texts: []TextPart{{Text: "end"}}},
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "end def"}}},
			}, wantErr: false},
		{name: "<prosody> rate", xml: `<speak><prosody rate="50%">aaa</prosody></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 2, Texts: []TextPart{{Text: "aaa"}}},
			}, wantErr: false},
		{name: "<prosody> volume", xml: `<speak><prosody volume="silent">aaa</prosody></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, VolumeChange: MinVolumeChange, Texts: []TextPart{{Text: "aaa"}}},
			}, wantErr: false},
		{name: "<prosody> volume", xml: `<speak><prosody volume="+10dB">aaa</prosody></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, VolumeChange: 10, Texts: []TextPart{{Text: "aaa"}}},
			}, wantErr: false},
		{name: "<prosody> volume", xml: `<speak><prosody volume="-5dB">aaa</prosody></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, VolumeChange: -5, Texts: []TextPart{{Text: "aaa"}}},
			}, wantErr: false},
		{name: "<prosody> pitch", xml: `<speak><prosody pitch="+10st">aaa</prosody></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, VolumeChange: 0, PitchChanges: []PitchChange{{Value: 10, Kind: PitchChangeSemitone}}, Texts: []TextPart{{Text: "aaa"}}},
			}, wantErr: false},
		{name: "<prosody> volume and rate", xml: `<speak><prosody rate="50%" volume="-5dB">aaa</prosody></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 2, VolumeChange: -5, Texts: []TextPart{{Text: "aaa"}}},
			}, wantErr: false},
		{name: "<prosody> inner", xml: `<speak><prosody rate="200%">
		<voice name="ooo">aaa
		<voice name="ooo1"><prosody rate="slow">aaa1</prosody></voice>
		end
		</voice></prosody>
		end def</speak>`,
			want: []Part{
				&Text{Voice: "ooo", Speed: 0.5, Texts: []TextPart{{Text: "aaa"}}},
				&Text{Voice: "ooo1", Speed: 0.75, Texts: []TextPart{{Text: "aaa1"}}},
				&Text{Voice: "ooo", Speed: 0.5, Texts: []TextPart{{Text: "end"}}},
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "end def"}}},
			}, wantErr: false},
		{name: "<prosody> inner volume and rate", xml: `<speak><prosody rate="200%" volume="-3dB">
		<voice name="ooo">aaa
		<voice name="ooo1"><prosody rate="slow" volume="-5dB">aaa1</prosody></voice>
		end
		</voice></prosody>
		end def</speak>`,
			want: []Part{
				&Text{Voice: "ooo", Speed: 0.5, VolumeChange: -3, Texts: []TextPart{{Text: "aaa"}}},
				&Text{Voice: "ooo1", Speed: 0.75, VolumeChange: -8, Texts: []TextPart{{Text: "aaa1"}}},
				&Text{Voice: "ooo", Speed: 0.5, VolumeChange: -3, Texts: []TextPart{{Text: "end"}}},
				&Text{Voice: "aa", Speed: 1, VolumeChange: 0, Texts: []TextPart{{Text: "end def"}}},
			}, wantErr: false},
		{name: "<prosody> inner volume, pitch and rate, ", xml: `
		<speak>
			<prosody rate="200%" volume="-3dB" pitch="+2st">
				<voice name="ooo">aaa
					<voice name="ooo1">
						<prosody rate="slow" volume="-5dB" pitch="+10%">
						aaa1
						</prosody>
					</voice>
					end
				</voice>
				<prosody pitch="+10Hz">
					<prosody pitch="+10%">
						aaa2
					</prosody>
				</prosody>
			</prosody>
			end def
		</speak>`,
			want: []Part{
				&Text{Voice: "ooo", Speed: 0.5, VolumeChange: -3, PitchChanges: []PitchChange{{Value: 2, Kind: PitchChangeSemitone}}, Texts: []TextPart{{Text: "aaa"}}},
				&Text{Voice: "ooo1", Speed: 0.75, VolumeChange: -8, PitchChanges: []PitchChange{{Value: 2, Kind: PitchChangeSemitone},
					{Value: 1.1, Kind: PitchChangeMultiplier}}, Texts: []TextPart{{Text: "aaa1"}}},
				&Text{Voice: "ooo", Speed: 0.5, VolumeChange: -3, PitchChanges: []PitchChange{{Value: 2, Kind: PitchChangeSemitone}}, Texts: []TextPart{{Text: "end"}}},
				&Text{Voice: "aa", Speed: 0.5, VolumeChange: -3, PitchChanges: []PitchChange{{Value: 2, Kind: PitchChangeSemitone},
					{Value: 10, Kind: PitchChangeHertz}, {Value: 1.1, Kind: PitchChangeMultiplier}}, Texts: []TextPart{{Text: "aaa2"}}},
				&Text{Voice: "aa", Speed: 1, VolumeChange: 0, PitchChanges: nil, Texts: []TextPart{{Text: "end def"}}},
			}, wantErr: false},
		{name: "<voice> map", xml: `<speak><voice name="ooo">aaa</voice></speak>`,
			vf: func(s string) (string, error) { return "ooo.v1", nil },
			want: []Part{
				&Text{Voice: "ooo.v1", Speed: 1, Texts: []TextPart{{Text: "aaa"}}},
			}, wantErr: false},
		{name: "<voice> map fail", xml: `<speak><voice name="ooo">aaa</voice></speak>`,
			vf:   func(s string) (string, error) { return "", errors.New("no voice") },
			want: []Part{}, wantErr: true},
		{name: "syntax error", xml: "<speak>olia</sss>", want: []Part{},
			wantErr: true},
		{name: "speak empty text", xml: "<speak></speak>   ", want: []Part{},
			wantErr: false},
		{name: "speak empty text", xml: "<speak></speak> \n\n\n  ", want: []Part{},
			wantErr: false},
		{name: "<w> ", xml: `<speak><intelektika:w acc="g{a/}li">gali</intelektika:w></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "gali", Accented: "g{a/}li"}}},
			}, wantErr: false},
		{name: "<w> joins", xml: `<speak>olia<intelektika:w acc="g{a/}li">gali</intelektika:w></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "olia"}, {Text: "gali", Accented: "g{a/}li"}}},
			}, wantErr: false},
		{name: "<w> several", xml: `<speak><intelektika:w acc="g{a/}li">gali</intelektika:w>
		<intelektika:w acc="{a/}li">gali</intelektika:w>
		</speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "gali", Accented: "g{a/}li"}, {Text: "gali", Accented: "{a/}li"}}},
			}, wantErr: false},
		{name: "<w> splits", xml: `<speak><intelektika:w acc="g{a/}li">gali</intelektika:w><p/><intelektika:w acc="g{a/}li">gali</intelektika:w>
			</speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "gali", Accented: "g{a/}li"}}},
				&Pause{Duration: pDuration},
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "gali", Accented: "g{a/}li"}}},
			}, wantErr: false},
		{name: "<voice> splits", xml: `<speak>text1<voice name="v1">gali</voice>text2</speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "text1"}}},
				&Text{Voice: "v1", Speed: 1, Texts: []TextPart{{Text: "gali"}}},
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "text2"}}},
			}, wantErr: false},
		{name: "joins comment", xml: `<speak>text1
			<!-- comment -->
			text2</speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "text1"}, {Text: "text2"}}},
			}, wantErr: false},
		{name: "fail <w> in <w>", xml: `<speak><intelektika:w acc="g{a/}li"><intelektika:w acc="g{a/}li">gali</intelektika:w></intelektika:w></speak>`,
			want: nil, wantErr: true},
		{name: "fail <w> no text", xml: `<speak><intelektika:w acc="g{a/}li"></intelektika:w></speak>`,
			want: nil, wantErr: true},
		{name: "fail <w> no acc", xml: `<speak><intelektika:w acc="">gali</intelektika:w></speak>`,
			want: nil, wantErr: true},
		{name: "fail <w> in <break>", xml: `<speak><break time="10s"><intelektika:w acc="g{a/}li">gali</intelektika:w></break></speak>`,
			want: nil, wantErr: true},
		{name: "parses sylls", xml: `<speak><intelektika:w acc="g{a/}li" syll="ga-li">gali</intelektika:w></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "gali", Accented: "g{a/}li", Syllables: "ga-li"}}},
			}, wantErr: false},
		{name: "parses OE", xml: `<speak><intelektika:w acc="ole" syll="o-le" user="O*l'E">olia</intelektika:w></speak>`,
			want: []Part{
				&Text{Voice: "aa", Speed: 1, Texts: []TextPart{{Text: "olia", Accented: "ole", Syllables: "o-le", UserOEPal: "O*l'E"}}},
			}, wantErr: false},
		{name: "fails OE", xml: `<speak><intelektika:w acc="ole" syll="o-le" user="OlEe">olia</intelektika:w></speak>`,
			want: nil, wantErr: true},
		{name: "fails sylls", xml: `<speak><intelektika:w acc="ole" syll="oo-le" user="OlE">olia</intelektika:w></speak>`,
			want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := &Text{Voice: "aa", Speed: 1}
			vf := vf
			if tt.vf != nil {
				vf = tt.vf
			}
			got, err := Parse(strings.NewReader(tt.xml), def, vf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("Parse() = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if !reflect.DeepEqual(got[i], tt.want[i]) {
					t.Errorf("Parse() = %v, want %v", got[i], tt.want[i])
				}
			}
		})
	}
}

func Test_getDuration(t *testing.T) {
	type args struct {
		tm  string
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{name: "time", args: args{tm: "10s", str: ""}, want: time.Second * 10, wantErr: false},
		{name: "time", args: args{tm: "10ms", str: ""}, want: time.Millisecond * 10, wantErr: false},
		{name: "time", args: args{tm: "1s250ms", str: ""}, want: time.Millisecond * 1250, wantErr: false},
		{name: "strength", args: args{tm: "", str: "none"}, want: time.Millisecond * 0, wantErr: false},
		{name: "strength", args: args{tm: "", str: "x-weak"}, want: time.Millisecond * 250, wantErr: false},
		{name: "strength", args: args{tm: "", str: "weak"}, want: time.Millisecond * 500, wantErr: false},
		{name: "strength", args: args{tm: "", str: "medium"}, want: time.Millisecond * 750, wantErr: false},
		{name: "strength", args: args{tm: "", str: "strong"}, want: time.Millisecond * 1000, wantErr: false},
		{name: "strength", args: args{tm: "", str: "x-strong"}, want: time.Millisecond * 1250, wantErr: false},
		{name: "strength fail", args: args{tm: "", str: "aaa"}, want: time.Millisecond * 0, wantErr: true},
		{name: "none", args: args{tm: "", str: ""}, want: time.Millisecond * 0, wantErr: true},
		{name: "time fail", args: args{tm: "xxx", str: ""}, want: time.Millisecond * 0, wantErr: true},
		{name: "both fail", args: args{tm: "xxx", str: "aaa"}, want: time.Millisecond * 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDuration(tt.args.tm, tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSpeed(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    float64
		wantErr bool
	}{
		{name: "percent", args: "10%", want: 2, wantErr: false},
		{name: "percent", args: "-40%", want: 2, wantErr: false},
		{name: "percent", args: "50%", want: 2, wantErr: false},
		{name: "percent", args: "75%", want: 1.5, wantErr: false},
		{name: "percent", args: "100%", want: 1, wantErr: false},
		{name: "percent", args: "120%", want: .9, wantErr: false},
		{name: "percent", args: "150%", want: .75, wantErr: false},
		{name: "percent", args: "200%", want: .5, wantErr: false},
		{name: "percent", args: "300%", want: .5, wantErr: false},
		{name: "percent fail", args: "300", want: 0, wantErr: true},
		{name: "percent fail", args: "aa%", want: 0, wantErr: true},
		{name: "rate", args: "x-slow", want: 2, wantErr: false},
		{name: "rate", args: "slow", want: 1.5, wantErr: false},
		{name: "rate", args: "medium", want: 1, wantErr: false},
		{name: "rate", args: "default", want: 1, wantErr: false},
		{name: "rate", args: "fast", want: .75, wantErr: false},
		{name: "rate", args: "x-fast", want: .5, wantErr: false},
		{name: "rate fail", args: "aaa x-slow", want: 0, wantErr: true},
		{name: "empty fail", args: "", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSpeed(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSpeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSpeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_okAccentedWord(t *testing.T) {
	tests := []struct {
		name string
		args string
		want bool
	}{
		{name: "word", args: "aaa", want: true},
		{name: "acc", args: "{a/}", want: true},
		{name: "empty", args: " ", want: false},
		{name: "empty", args: "", want: false},
		{name: "long", args: strings.Repeat("a", 50), want: false},
		{name: "wrong acc", args: "{a-}", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := okAccentedWord(tt.args); got != tt.want {
				t.Errorf("okAccentedWord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_clearSylls(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{s: ""}, want: ""},
		{name: "none", args: args{s: "olia"}, want: "olia"},
		{name: "clears", args: args{s: "o-li-a"}, want: "olia"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clearSylls(tt.args.s); got != tt.want {
				t.Errorf("clearSylls() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_clearUserOE(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{s: ""}, want: ""},
		{name: "none", args: args{s: "Olia"}, want: "Olia"},
		{name: "clears", args: args{s: "Ol'iab*a"}, want: "Oliaba"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clearUserOE(tt.args.s); got != tt.want {
				t.Errorf("clearUserOE() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getVolume(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    float64
		wantErr bool
	}{
		{name: "silent", str: "silent", want: MinVolumeChange, wantErr: false},
		{name: "x-soft", str: "x-soft", want: -6, wantErr: false},
		{name: "soft", str: "soft", want: -3, wantErr: false},
		{name: "medium", str: "medium", want: 0, wantErr: false},
		{name: "loud", str: "loud", want: 3, wantErr: false},
		{name: "x-loud", str: "x-loud", want: 6, wantErr: false},
		{name: "default", str: "default", want: 0, wantErr: false},
		{name: "plus dB", str: "+5dB", want: 5, wantErr: false},
		{name: "plus dB", str: "+0.5dB", want: 0.5, wantErr: false},
		{name: "minus dB", str: "-3dB", want: -3, wantErr: false},
		{name: "invalid string", str: "aaa", want: 0, wantErr: true},
		{name: "invalid plus", str: "+dB", want: 0, wantErr: true},
		{name: "invalid minus", str: "-dB", want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getVolume(tt.str)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("getVolume() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("getVolume() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("getVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPitch(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    PitchChange
		wantErr bool
	}{
		{name: "semitone plus", str: "+5st", want: PitchChange{Value: 5, Kind: PitchChangeSemitone}, wantErr: false},
		{name: "semitone minus", str: "-3st", want: PitchChange{Value: -3, Kind: PitchChangeSemitone}, wantErr: false},
		{name: "hertz plus", str: "+10Hz", want: PitchChange{Value: 10, Kind: PitchChangeHertz}, wantErr: false},
		{name: "hertz minus", str: "-4Hz", want: PitchChange{Value: -4, Kind: PitchChangeHertz}, wantErr: false},
		{name: "percent plus", str: "+20%", want: PitchChange{Value: 1.2, Kind: PitchChangeMultiplier}, wantErr: false},
		{name: "percent minus", str: "-50%", want: PitchChange{Value: 0.5, Kind: PitchChangeMultiplier}, wantErr: false},
		{name: "invalid string", str: "aaa", want: PitchChange{}, wantErr: true},
		{name: "invalid semitone", str: "+st", want: PitchChange{}, wantErr: true},
		{name: "invalid hertz", str: "-Hz", want: PitchChange{}, wantErr: true},
		{name: "invalid percent", str: "+%", want: PitchChange{}, wantErr: true},
		{name: "x-low", str: "x-low", want: PitchChange{Value: 0.55, Kind: PitchChangeMultiplier}, wantErr: false},
		{name: "low", str: "low", want: PitchChange{Value: 0.8, Kind: PitchChangeMultiplier}, wantErr: false},
		{name: "medium", str: "medium", want: PitchChange{Value: 1, Kind: PitchChangeMultiplier}, wantErr: false},
		{name: "high", str: "high", want: PitchChange{Value: 1.2, Kind: PitchChangeMultiplier}, wantErr: false},
		{name: "x-high", str: "x-high", want: PitchChange{Value: 1.45, Kind: PitchChangeMultiplier}, wantErr: false},
		{name: "default", str: "default", want: PitchChange{Value: 1, Kind: PitchChangeMultiplier}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getPitch(tt.str)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("getPitch() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("getPitch() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("getPitch() = %v, want %v", got, tt.want)
			}
		})
	}
}
