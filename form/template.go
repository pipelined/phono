package form

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
        a {
            color:inherit;
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
        .file {
            margin-bottom: 20px;
        }
        .container {
            width: 1000px;
            margin-right: auto;
            margin-left: auto;
        }
        .outputs {
            margin-bottom: 20px;
            display: block;
        }
        .output-options {
            display: none;
        }
        .mp3-bit-rate-mode-options{
            display: none;
        }
        .mp3-quality {
            display: inline;
        }
        .option {
            margin-right: 7px;
        }
        .footer{
            position: fixed;
            padding-top: 15px;
            padding-bottom: 15px;
            bottom: 0;
        }
        #output-format {
            display: none;
        }
        #input-file {
            display: none;
        }
        #input-file-label {
            cursor: pointer;
            padding:0!important;
            border-bottom:1px solid #444; 
        }
    </style>
    <script type="text/javascript">
        document.addEventListener('DOMContentLoaded', function(event) {
            document.getElementById('convert').reset();
        });
        function getFileName(id) {
            var filePath = document.getElementById(id).value;
            return filePath.substr(filePath.lastIndexOf('\\') + 1);
        }
        function displayClass(className, mode) {
            var elements = document.getElementsByClassName(className);
            for (var i = 0, ii = elements.length; i < ii; i++) {
                elements[i].style.display = mode;
            };
        }
        function displayId(id, mode){
            document.getElementById(id).style.display = mode;
        }
        function onInputFileChange(){
            document.getElementById('input-file-label').innerHTML = getFileName('input-file');
            displayClass('input-file-label', 'inline');
            displayId('output-format', 'inline');
        }
		function onOutputFormatChange(el){
            displayClass('output-options', 'none');
            // need to cut the dot
        	displayId(el.value.slice(1)+'-options', 'inline');
        	displayClass('submit', 'block');
        }
        function onMp3BitRateModeChange(el){
        	displayClass('mp3-bit-rate-mode-options', 'none');
        	var selectedOptions = 'mp3-'+el.options[el.selectedIndex].id+'-options';
        	displayClass(selectedOptions, 'inline');
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
    <div class="container">
        <h2>phono convert</h1>
        <form id="convert" enctype="multipart/form-data" method="post">
        <div class="file">
            <input id="input-file" type="file" name="input-file" accept="{{.Accept}}" onchange="onInputFileChange()"/>
            <label id="input-file-label" for="input-file">select file</label>
        </div>
        <div class="outputs">
            <div id="output-format" class="option">
                format 
                <select name="format" onchange="onOutputFormatChange(this)">
                    <option hidden disabled selected value>select</option>
                    {{range $key, $value := .OutFormats}}
                        <option id="{{ $value }}" value="{{ $key }}">{{ $value }}</option>
                    {{end}}
                </select>
            </div>
            <div id="wav-options" class="output-options">
                bit depth
                <select name="wav-bit-depth" class="option">
                    <option hidden disabled selected value>select</option>
                    {{range $key, $value := .WavOptions.BitDepths}}
                        <option value="{{ printf "%d" $key }}">{{ $key }}</option>
                    {{end}}
                </select>
            </div>
            <div id="mp3-options" class="output-options">
                channel mode
                <select name="mp3-channel-mode" class="option">
                    <option hidden disabled selected value>select</option>
                    {{range $key, $value := .Mp3Options.ChannelModes}}
                        <option value="{{ printf "%d" $key }}">{{ $key }}</option>
                    {{end}}
                </select>
                bit rate mode
                <select id="mp3-bit-rate-mode" class="option" name="mp3-bit-rate-mode" onchange="onMp3BitRateModeChange(this)">
                    <option hidden disabled selected value>select</option>
                    {{range $key, $value := .Mp3Options.BitRateModes}}
                        <option id="{{ $key }}" value="{{ printf "%d" $key }}">{{ $key }}</option>
                    {{end}}
                </select>
                <div class="mp3-bit-rate-mode-options mp3-{{ .Mp3Options.ABR }}-options mp3-{{ .Mp3Options.CBR }}-options">
                    bit rate [8-320]
                    <input type="text" class="option" name="mp3-bit-rate" maxlength="3" size="3">
                </div> 
                <div class="mp3-bit-rate-mode-options mp3-{{ .Mp3Options.VBR }}-options">
                    vbr quality [0-9]
                    <input type="text" class="option" name="mp3-vbr-quality" maxlength="1" size="3">
                </div>
                <div class="mp3-quality">
                    <input type="checkbox" id="mp3-use-quality" name="mp3-use-quality" value="true" onchange="onMp3UseQUalityChange(this)">quality
                    <div id="mp3-quality-value" class="mp3-quality" style="visibility:hidden">
                        [0-9]
                        <input type="text" class="option" name="mp3-quality" maxlength="1" size="3">
                    </div> 
                </div>
            </div>
        </div>
        </form>
        <div class="submit" style="display:none">
            <button type="button" onclick="onSubmitClick()">convert</button> 
        </div>
        <div class="footer">
            <div class="container">
            made with <a href="https://github.com/pipelined/pipe" target="_blank">pipe</a>
            </div>
        </div>
    </div>
</body>
</html>
`
