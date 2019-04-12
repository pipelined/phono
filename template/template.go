package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/pipelined/mp3"
	"github.com/pipelined/signal"
	"github.com/pipelined/wav"
)

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

// accept generates string for accept html form attribute.
func accept(extFns ...extensionsFunc) string {
	var str strings.Builder
	for _, fn := range extFns {
		for _, ext := range fn() {
			str.WriteString(ext + ",")
		}
	}
	return str.String()
}

// outFormats returns map of extensions without dots.
func outFormats(exts ...string) map[string]string {
	m := make(map[string]string)
	for _, ext := range exts {
		m[ext] = ext[1:]
	}
	return m
}

type extensionsFunc func() []string

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
                <option value="{{ printf "%d" $key }}">{{ $key }}</option>
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
