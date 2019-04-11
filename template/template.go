package template

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/pipelined/mp3"
	"github.com/pipelined/phono/convert"
	"github.com/pipelined/pipe"
	"github.com/pipelined/signal"
	"github.com/pipelined/wav"
)

// ConvertForm provides user interaction via http form.
type ConvertForm struct{}

// convertData provides a data for convert form, so user can define conversion parameters.
type convertData struct {
	Accept     string
	OutFormats map[string]string
	WavOptions wavOptions
	Mp3Options mp3Options
}

// wavOptions is a struct of wav format options that are available for conversion.
type wavOptions struct {
	BitDepths map[signal.BitDepth]struct{}
}

// mp3Options is a struct of mp3 format options that are available for conversion.
type mp3Options struct {
	VBR           mp3.BitRateMode
	ABR           mp3.BitRateMode
	CBR           mp3.BitRateMode
	BitRateModes  map[mp3.BitRateMode]struct{}
	ChannelModes  map[mp3.ChannelMode]struct{}
	DefineQuality bool
}

var (
	// ConvertFormData is the serialized convert form with values.
	convertFormData []byte
)

// ErrUnsupportedConfig is returned when unsupported configuraton passed.
type ErrUnsupportedConfig string

// Error returns error message.
func (e ErrUnsupportedConfig) Error() string {
	return string(e)
}

// init generates serialized convert form data which is then used during runtime.
func init() {
	convertTemplate := template.Must(template.New("convert").Parse(convertHTML))

	convertFormValues := convertData{
		Accept:     accept(mp3.Extensions, wav.Extensions),
		OutFormats: outFormats(mp3.DefaultExtension, wav.DefaultExtension),
		WavOptions: wavOptions{
			BitDepths: wav.Supported.BitDepths,
		},
		Mp3Options: mp3Options{
			VBR:          mp3.VBR,
			ABR:          mp3.ABR,
			CBR:          mp3.CBR,
			BitRateModes: mp3.Supported.BitRateModes,
			ChannelModes: mp3.Supported.ChannelModes,
		},
	}
	var b bytes.Buffer
	if err := convertTemplate.Execute(&b, convertFormValues); err != nil {
		panic(fmt.Sprintf("Failed to parse convert template: %v", err))
	}
	convertFormData = b.Bytes()
}

// accept generates
func accept(extFns ...extensionsFunc) string {
	var str strings.Builder
	for _, fn := range extFns {
		for _, ext := range fn() {
			str.WriteString(ext + ",")
		}
	}
	return str.String()
}

func outFormats(exts ...string) map[string]string {
	m := make(map[string]string)
	for _, ext := range exts {
		m[ext] = ext[1:]
	}
	return m
}

// Data returns serialized form data, ready to be served.
func (ConvertForm) Data() []byte {
	return convertFormData
}

// Format parses input format from http request.
func (ConvertForm) Format(r *http.Request) string {
	return path.Base(r.URL.Path)
}

// Parse form data into output config.
func (ConvertForm) Parse(r *http.Request) (convert.SinkBuilder, error) {
	ext := r.FormValue("format")
	switch ext {
	case wav.DefaultExtension:
		return parseWavConfig(r)
	case mp3.DefaultExtension:
		return parseMp3Config(r)
	default:
		return nil, ErrUnsupportedConfig(fmt.Sprintf("Unsupported format: %v", ext))
	}
}

// Pump parses http request and returns pump.
func (ConvertForm) Pump(r *http.Request) (pipe.Pump, io.Closer, error) {
	f, handler, err := r.FormFile("input-file")
	if err != nil {
		return nil, nil, fmt.Errorf("Invalid file: %v", err)
	}
	switch {
	case hasExtension(handler.Filename, wav.Extensions):
		return &wav.Pump{ReadSeeker: f}, f, nil
	case hasExtension(handler.Filename, mp3.Extensions):
		return &mp3.Pump{Reader: f}, f, nil
	default:
		extErr := fmt.Errorf("File has unsupported extension: %v", handler.Filename)
		if err = f.Close(); err != nil {
			return nil, nil, fmt.Errorf("%s \nFailed close form file: %v", extErr, err)
		}
		return nil, nil, extErr
	}
}

// Source is the input for convertation.
type Source interface {
	io.Reader
	io.Seeker
	io.Closer
}

type extensionsFunc func() []string

func hasExtension(fileName string, fn extensionsFunc) bool {
	for _, ext := range fn() {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}
	return false
}

func parseWavConfig(r *http.Request) (*wav.SinkBuilder, error) {
	// try to get bit depth
	bitDepth, err := parseIntValue(r, "wav-bit-depth", "bit depth")
	if err != nil {
		return nil, err
	}

	return &wav.SinkBuilder{BitDepth: signal.BitDepth(bitDepth)}, nil
}

func parseMp3Config(r *http.Request) (*mp3.SinkBuilder, error) {
	// try to get bit rate mode
	bitRateMode, err := parseIntValue(r, "mp3-bit-rate-mode", "bit rate mode")
	if err != nil {
		return nil, err
	}

	// try to get channel mode
	channelMode, err := parseIntValue(r, "mp3-channel-mode", "channel mode")
	if err != nil {
		return nil, err
	}

	if mp3.BitRateMode(bitRateMode) == mp3.VBR {
		// try to get vbr quality
		vbrQuality, err := parseIntValue(r, "mp3-vbr-quality", "vbr quality")
		if err != nil {
			return nil, err
		}
		return &mp3.SinkBuilder{
			BitRateMode: mp3.VBR,
			ChannelMode: mp3.ChannelMode(channelMode),
			VBRQuality:  vbrQuality,
		}, nil
	}

	// try to get bitrate
	bitRate, err := parseIntValue(r, "mp3-bit-rate", "bit rate")
	if err != nil {
		return nil, err
	}
	return &mp3.SinkBuilder{
		BitRateMode: mp3.BitRateMode(bitRateMode),
		ChannelMode: mp3.ChannelMode(channelMode),
		BitRate:     bitRate,
	}, nil
}

// parseIntValue parses value of key provided in the html form.
// Returns error if value is not provided or cannot be parsed as int.
func parseIntValue(r *http.Request, key, name string) (int, error) {
	str := r.FormValue(key)
	if str == "" {
		return 0, ErrUnsupportedConfig(fmt.Sprintf("Please provide %s", name))
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, ErrUnsupportedConfig(fmt.Sprintf("Failed parsing %s %s: %v", name, str, err))
	}
	return val, nil
}

// func outFormats() map[string]string {}

const convertHTML = `
<html>
<head>
    <style>
        * {
            font-family: Verdana;
        }
        form {
            margin: 0;
        }
        button {
            background:none!important;
            color:inherit;
            border:none; 
            padding:0!important;
            font: inherit;
            border-bottom:1px solid #444; 
            cursor: pointer;
        }
        #input-file-label {
            cursor: pointer;
            padding:0!important;
            border-bottom:1px solid #444; 
        }
        #mp3-quality {
            padding-bottom: 10px;
        }
    </style>
    <script type="text/javascript">
        document.addEventListener("DOMContentLoaded", function(event) {
            document.getElementById("convert").reset();
        });
        function getFileName(id) {
            var filePath = document.getElementById(id).value;
            return filePath.substr(filePath.lastIndexOf('\\') + 1);
        }
        function displayClass(className, display) {
            var elements = document.getElementsByClassName(className);
            for (var i = 0, ii = elements.length; i < ii; i++) {
                elements[i].style.display = display ? '' : 'none';
            };
        }
        function displayId(id, mode){
            document.getElementById(id).style.display = mode;
        }
        function onInputFileChange(){
            document.getElementById('input-file-label').innerHTML = getFileName('input-file');
            displayClass('input-file-label', true);
            displayId('output-format', "");
        }
		function onOutputFormatsClick(el){
        	displayClass('output-options', false);
        	displayId(el.id+'-options', "");
        	displayId('submit', "");
        }
        function onMp3BitRateModeChange(el){
        	displayClass('mp3-bit-rate-mode-options', false);
        	var selectedOptions = 'mp3-'+el.options[el.selectedIndex].id+'-options';
        	displayClass(selectedOptions, true);
        }
        function onMp3UseQUalityChange(el){
            if (el.checked) {
                document.getElementById('mp3-quality-value').style.visibility = "";
            } else {
                document.getElementById('mp3-quality-value').style.visibility = "hidden";
            }
        }
        function onSubmitClick(){
            var fileName = getFileName('input-file')
            var ext = fileName.split('.')[1];
            var convert = document.getElementById('convert');
            convert.action = ext;
            convert.submit();
        }
    </script> 
</head>
<body>
    <h2>phono convert http</h1>
    <form id="convert" enctype="multipart/form-data" method="post">
    <div id="file">
        <input id="input-file" type="file" name="input-file" accept="{{.Accept}}" style="display:none" onchange="onInputFileChange()"/>
        <label id="input-file-label" for="input-file">select file</label>
    </div>
    <div id="output-format" style="display:none">
        output 
        {{range $key, $value := .OutFormats}}
            <input type="radio" id="{{ $value }}" value="{{ $key }}" name="format" class="output-formats" onclick="onOutputFormatsClick(this)">
            <label for="{{ $value }}">{{ $value }}</label>
        {{end}}
    <br>
    </div>
    <div id="wav-options" class="output-options" style="display:none">
        bit depth
        <select name="wav-bit-depth">
            <option hidden disabled selected value>select</option>
            {{range $key, $value := .WavOptions.BitDepths}}
                <option value="{{ printf "%d" $key }}">{{ $key }}</option>
            {{end}}
        </select>
    <br>
    </div>
    <div id="mp3-options" class="output-options" style="display:none">
        channel mode
        <select name="mp3-channel-mode">
            <option hidden disabled selected value>select</option>
            {{range $key, $value := .Mp3Options.ChannelModes}}
                <option value="{{ $key }}">{{ $key }}</option>
            {{end}}
        </select>
        <br>
        bit rate mode
        <select id="mp3-bit-rate-mode" name="mp3-bit-rate-mode" onchange="onMp3BitRateModeChange(this)">
            <option hidden disabled selected value>select</option>
            {{range $key, $value := .Mp3Options.BitRateModes}}
                <option id="{{ $key }}" value="{{ printf "%d" $key }}">{{ $key }}</option>
            {{end}}
        </select>
        <br>
        <div class="mp3-bit-rate-mode-options mp3-{{ .Mp3Options.ABR }}-options mp3-{{ .Mp3Options.CBR }}-options" style="display:none">
            bit rate [8-320]
            <input type="text" name="mp3-bit-rate" maxlength="3" size="3">
        </div> 
        <div class="mp3-bit-rate-mode-options mp3-{{ .Mp3Options.VBR }}-options" style="display:none">
            vbr quality [0-9]
            <input type="text" name="mp3-vbr-quality" maxlength="1" size="3">
        </div>
        <div id="mp3-quality">
            <input type="checkbox" id="mp3-use-quality" name="mp3-use-quality" value="true" onchange="onMp3UseQUalityChange(this)">quality
            <div id="mp3-quality-value" style="display:inline;visibility:hidden">
                [0-9]
                <input type="text" name="mp3-quality" maxlength="1" size="3">
            </div>
            <br>  
        </div>
    </div>
    </form>
    <button id="submit" type="button" style="display:none" onclick="onSubmitClick()">convert</button> 
</body>
</html>
`
