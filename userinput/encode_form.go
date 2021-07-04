package userinput

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	"pipelined.dev/audio/fileformat"

	"pipelined.dev/phono/encode"
)

var (
	errInputFormat  = errors.New("unsupported input format")
	errOutputFormat = errors.New("unsupported output format")
)

var formTemplate = template.Must(template.New("encode").Parse(encodeHTML))

// FormFileKey is the id of the file userinput in the HTML form.
const FormFileKey = "form-file"

type (
	// Limits for user-provided input files.
	Limits map[*fileformat.Format]int64

	// EncodeForm provides user interaction via http form.
	EncodeForm struct {
		buf    bytes.Buffer
		limits Limits
	}

	// templateData provides a data for encode form template, so user can
	// define conversion parameters.
	templateData struct {
		Accept     string
		OutFormats []string
		WAV        interface{}
		MP3        interface{}
		MaxSizes   map[string]int64
	}
)

// NewEncodeForm creates new form with provided limits.
func NewEncodeForm(limits Limits) EncodeForm {
	var buf bytes.Buffer
	err := formTemplate.Execute(&buf, templateData{
		MaxSizes: limits.maxSizes(),
		Accept: strings.Join(
			inputExtensions(
				fileformat.WAV(),
				fileformat.MP3(),
				fileformat.FLAC(),
			),
			", "),
		OutFormats: outputExtensions(
			fileformat.WAV(),
			fileformat.MP3(),
		),
		WAV: WAV,
		MP3: MP3,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to parse encode template: %v", err))
	}
	return EncodeForm{
		buf:    buf,
		limits: limits,
	}
}

// Bytes returns serialized form, ready to be served.
func (f EncodeForm) Bytes() []byte {
	return f.buf.Bytes()
}

// Parse returns the data provided by the user via submitted form.
func (f EncodeForm) Parse(r *http.Request) (encode.FormData, error) {
	inputFormat := fileformat.FormatByPath(r.URL.Path)
	if inputFormat == nil {
		return encode.FormData{}, errInputFormat
	}
	// get max size for the format
	maxSize := f.inputMaxSize(inputFormat)
	if maxSize > 0 {
		// check if limit is defined
		if maxSize > 0 {
			r.Body = http.MaxBytesReader(nil, r.Body, maxSize)
		}
	}
	// check max size
	if err := r.ParseMultipartForm(maxSize); err != nil {
		return encode.FormData{}, err
	}

	file, _, err := r.FormFile(FormFileKey)
	if err != nil {
		return encode.FormData{}, err
	}

	// parse sink and validate parameters
	sink, outputFormat, err := parseOutput(r.MultipartForm.Value)
	if err != nil {
		return encode.FormData{}, err
	}

	return encode.FormData{
		Input: encode.Input{
			Format: inputFormat,
			File:   file,
		},
		Output: encode.Output{
			Format: outputFormat,
			Sink:   sink,
		},
	}, nil
}

func inputExtensions(formats ...*fileformat.Format) []string {
	result := make([]string, 0, len(formats))
	for i := range formats {
		result = append(result, formats[i].Extensions()...)
	}
	return result
}

// outFormats maps the extensions with values without dots.
func outputExtensions(formats ...*fileformat.Format) []string {
	result := make([]string, 0, len(formats))
	for i := range formats {
		result = append(result, formats[i].DefaultExtension())
	}
	return result
}

func (l Limits) maxSizes() map[string]int64 {
	m := make(map[string]int64)
	for format, limit := range l {
		for _, ext := range format.Extensions() {
			m[ext] = limit
		}
	}
	return m
}

// inputMaxSize of file from http request.
func (f EncodeForm) inputMaxSize(format *fileformat.Format) int64 {
	return f.limits[format]
}

// ParseForm provided via form.
// This function should return extensions, sinkbuilder
func parseOutput(formData url.Values) (Sink, *fileformat.Format, error) {
	formatString := strings.ToLower(formData.Get("format"))
	format := fileformat.FormatByPath(formatString)
	var (
		sink Sink
		err  error
	)
	switch format {
	case fileformat.WAV():
		sink, err = parseWAVSink(formData)
	case fileformat.MP3():
		sink, err = parseMP3Sink(formData)
	default:
		return nil, nil, fmt.Errorf("Unsupported format: %v", formatString)
	}
	return sink, format, err
}

func parseWAVSink(data url.Values) (Sink, error) {
	// try to get bit depth
	bitDepth, err := parseIntValue(data, "wav-bit-depth", "bit depth")
	if err != nil {
		return nil, err
	}
	return WAV.Sink(bitDepth)
}

func parseMP3Sink(data url.Values) (Sink, error) {
	// try to get channel mode
	channelMode, err := parseIntValue(data, "mp3-channel-mode", "channel mode")
	if err != nil {
		return nil, err
	}

	var bitRate int
	// try to get bit rate mode
	bitRateMode := data.Get("mp3-bit-rate-mode")
	switch bitRateMode {
	case MP3.VBR:
		// try to get vbr quality
		bitRate, err = parseIntValue(data, "mp3-vbr-quality", "vbr quality")
		if err != nil {
			return nil, err
		}
	case MP3.CBR, MP3.ABR:
		// try to get bitrate
		bitRate, err = parseIntValue(data, "mp3-bit-rate", "bit rate")
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Unsupported bit rate mode: %v", bitRateMode)
	}

	// try to get mp3 quality
	useQuality, err := parseBoolValue(data, "mp3-use-quality", "mp3 quality")
	if err != nil {
		return nil, err
	}
	var quality int
	if useQuality {
		quality, err = parseIntValue(data, "mp3-quality", "mp3 quality")
		if err != nil {
			return nil, err
		}
	}

	return MP3.Sink(bitRateMode, bitRate, channelMode, useQuality, quality)
}

// parseIntValue parses value of key provided in the html form. Returns
// error if value is not provided or cannot be parsed as int.
func parseIntValue(data url.Values, key, name string) (int, error) {
	str := data.Get(key)
	if str == "" {
		return 0, fmt.Errorf("%s not provided", name)
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("Failed parsing %s %s: %v", name, str, err)
	}
	return val, nil
}

// parseBoolValue parses value of key provided in the html form. Returns
// false if value is not provided. Returns error when cannot be parsed as
// bool.
func parseBoolValue(data url.Values, key, name string) (bool, error) {
	str := data.Get(key)
	if str == "" {

		return false, nil
	}

	val, err := strconv.ParseBool(str)
	if err != nil {
		return false, fmt.Errorf("Failed parsing %s %s: %v", name, str, err)
	}
	return val, nil
}

const encodeHTML = `
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
        #output-format-block {
            display: none;
        }
        #form-file {
            display: none;
        }
        #form-file-label {
            cursor: pointer;
            padding:0!important;
            border-bottom:1px solid #444;
        }
    </style>
    <script type="text/javascript">
        const fileId = 'form-file';
        const accept = '{{ .Accept }}';
        function getFile() {
            return document.getElementById(fileId);
        }
        function getFileName(file) {
            var filePath = file.value;
            return filePath.substr(filePath.lastIndexOf('\\') + 1);
        }
        function getFileExtension(fileName) {
            return '.'.concat(fileName.split('.')[1]);
        }
        function humanFileSize(size) {
            var i = size == 0 ? 0 : Math.floor(Math.log(size) / Math.log(1024));
            return (size / Math.pow(1024, i)).toFixed(2) * 1 + ' ' + ['B', 'kB', 'MB', 'GB', 'TB'][i];
        };
        function displayClass(className, mode) {
            var elements = document.getElementsByClassName(className);
            for (var i = 0, ii = elements.length; i < ii; i++) {
                elements[i].style.display = mode;
            };
        }
        function displayId(id, mode){
            document.getElementById(id).style.display = mode;
        }
        document.addEventListener('DOMContentLoaded', function(event) {
            document.getElementById('encode').reset();
            // base form handlers
            document.getElementById('form-file').addEventListener('change', onInputFileChange);
            document.getElementById('output-format').addEventListener('change', onOutputFormatChange);
            document.getElementById('submit-button').addEventListener('click', onSubmitClick);
            // mp3 handlers
            document.getElementById('mp3-bit-rate-mode').addEventListener('change', onMp3BitRateModeChange);
            document.getElementById('mp3-use-quality').addEventListener('click', onMp3UseQUalityChange);
        });
        function onInputFileChange(){
            var fileName = getFileName(getFile());
            document.getElementById('form-file-label').innerHTML = fileName;
            var ext = getFileExtension(fileName);
            if (accept.indexOf(ext) < 0) {
                alert('Only files with following extensions are allowed: {{.Accept}}')
                return;
            }
            displayClass('form-file-label', 'inline');
            displayId('output-format-block', 'inline');
        }
		function onOutputFormatChange(){
            displayClass('output-options', 'none');
            // need to cut the dot
        	displayId(this.value.slice(1)+'-options', 'inline');
        	displayClass('submit', 'block');
        }
        function onMp3BitRateModeChange(){
        	displayClass('mp3-bit-rate-mode-options', 'none');
        	var selectedOptions = 'mp3-'+this.options[this.selectedIndex].id+'-options';
        	displayClass(selectedOptions, 'inline');
        }
        function onMp3UseQUalityChange(){
            if (this.checked) {
                document.getElementById('mp3-quality-value').style.visibility = '';
            } else {
                document.getElementById('mp3-quality-value').style.visibility = 'hidden';
            }
        }
        function onSubmitClick(){
            var encode = document.getElementById('encode');
            var file = getFile();
            var ext = getFileExtension(getFileName(file));
            encode.action = ext;
            var size = file.files[0].size;
            switch (ext) {
            {{ range $ext, $maxSize := .MaxSizes }}
            case '{{$ext}}':
                if ({{ $maxSize }} > 0 && {{ $maxSize }} < size) {
                    alert('File is too big. Maximum allowed size: '.concat(humanFileSize({{ $maxSize }})))
                    return;
                }
                break;
            {{ end }}
            }
            encode.submit();
        }
    </script>
</head>
<body>
    <div class="container">
        <h2>phono encode</h1>
        <form id="encode" enctype="multipart/form-data" method="post">
        <div class="file">
            <input id="form-file" type="file" name="form-file" accept="{{.Accept}}"/>
            <label id="form-file-label" for="form-file">select file</label>
        </div>
        <div class="outputs">
            <div id="output-format-block" class="option">
                format
                <select id="output-format" name="format">
                    <option hidden disabled selected value>select</option>
                    {{range $value := .OutFormats}}
                        <option id="{{ $value }}" value="{{ $value }}">{{ $value }}</option>
                    {{end}}
                </select>
            </div>
            <div id="wav-options" class="output-options">
                bit depth
                <select name="wav-bit-depth" class="option">
                    <option hidden disabled selected value>select</option>
                    {{range $key, $value := .WAV.BitDepths}}
                        <option value="{{ printf "%d" $key }}">{{ $key }}</option>
                    {{end}}
                </select>
            </div>
            <div id="mp3-options" class="output-options">
                channel mode
                <select name="mp3-channel-mode" class="option">
                    <option hidden disabled selected value>select</option>
                    {{range $key, $value := .MP3.ChannelModes}}
                        <option value="{{ printf "%d" $key }}">{{ $key }}</option>
                    {{end}}
                </select>
                bit rate mode
                <select id="mp3-bit-rate-mode" class="option" name="mp3-bit-rate-mode">
                    <option hidden disabled selected value>select</option>
                    <option id="{{ .MP3.VBR  }}" value="{{ .MP3.VBR }}">{{ .MP3.VBR }}</option>
                    <option id="{{ .MP3.CBR  }}" value="{{ .MP3.CBR }}">{{ .MP3.CBR }}</option>
                    <option id="{{ .MP3.ABR  }}" value="{{ .MP3.ABR }}">{{ .MP3.ABR }}</option>
                </select>
                <div class="mp3-bit-rate-mode-options mp3-{{ .MP3.ABR }}-options mp3-{{ .MP3.CBR }}-options">
                    bit rate [{{ .MP3.MinBitRate }}-{{ .MP3.MaxBitRate }}]
                    <input type="text" class="option" name="mp3-bit-rate" maxlength="3" size="3">
                </div>
                <div class="mp3-bit-rate-mode-options mp3-{{ .MP3.VBR }}-options">
                    vbr quality [{{ .MP3.MinVBR }}-{{ .MP3.MaxVBR }}]
                    <input type="text" class="option" name="mp3-vbr-quality" maxlength="1" size="3">
                </div>
                <div class="mp3-quality">
                    <input type="checkbox" id="mp3-use-quality" name="mp3-use-quality" value="true">quality
                    <div id="mp3-quality-value" class="mp3-quality" style="visibility:hidden">
                        [{{ .MP3.MinQuality }}-{{ .MP3.MaxQuality }}]
                        <input type="text" class="option" name="mp3-quality" maxlength="1" size="3">
                    </div>
                </div>
            </div>
        </div>
        </form>
        <div class="submit" style="display:none">
            <button id="submit-button" type="button">encode</button>
        </div>
        <div class="footer">
            <div class="container">
            powered by <a href="https://github.com/pipelined/pipe" target="_blank">pipe</a>
            </div>
        </div>
    </div>
</body>
</html>
`
