package form

import (
	"bytes"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/pipelined/phono/file"
)

// encodeData provides a data for encode form, so user can define conversion parameters.
type encodeData struct {
	Accept     string
	OutFormats map[string]string
	Wav        interface{}
	Mp3        interface{}
	MaxSizes   map[string]int64
}

const (
	// FileKey is the id of the file input in the HTML form.
	FileKey = "input-file"
)

func maxSizes(wavMaxSize, mp3MaxSize int64) map[string]int64 {
	m := make(map[string]int64)
	for ext := range file.Mp3.Extensions {
		m[ext] = mp3MaxSize
	}
	for ext := range file.Wav.Extensions {
		m[ext] = wavMaxSize
	}
	return m
}

func mergeExtensions(exts ...map[string]struct{}) []string {
	result := make([]string, 0)
	for _, m := range exts {
		for ext := range m {
			result = append(result, ext)
		}
	}
	return result
}

// outFormats maps the extensions with values without dots.
func outFormats(exts ...string) map[string]string {
	m := make(map[string]string)
	for _, ext := range exts {
		m[ext] = ext[1:]
	}
	return m
}

var (
	extensions = mergeExtensions(
		file.Wav.Extensions,
		file.Mp3.Extensions,
		file.Flac.Extensions,
	)

	encodeForm = encodeData{
		Accept: strings.Join(extensions, ", "),
		OutFormats: outFormats(
			file.Wav.DefaultExtension,
			file.Mp3.DefaultExtension,
		),
		Wav: file.Wav,
		Mp3: file.Mp3,
	}

	encodeTmpl = template.Must(template.New("encode").Parse(encodeHTML))
)

// Encode provides user interaction via http form.
type Encode struct {
	WavMaxSize  int64
	Mp3MaxSize  int64
	FlacMaxSize int64
}

// Data returns serialized form data, ready to be served.
func (c Encode) Data() []byte {
	d := encodeForm
	d.MaxSizes = maxSizes(c.WavMaxSize, c.Mp3MaxSize)

	var b bytes.Buffer
	if err := encodeTmpl.Execute(&b, d); err != nil {
		panic(fmt.Sprintf("Failed to parse encode template: %v", err))
	}
	return b.Bytes()
}

// InputMaxSize of file from http request.
func (c Encode) InputMaxSize(url string) (int64, error) {
	ext := strings.ToLower(path.Base(url))
	switch ext {
	case file.Mp3.DefaultExtension:
		return c.Mp3MaxSize, nil
	case file.Wav.DefaultExtension:
		return c.WavMaxSize, nil
	case file.Flac.DefaultExtension:
		return c.FlacMaxSize, nil
	default:
		return 0, fmt.Errorf("Format %s not supported", ext)
	}
}

// FileKey returns a name of form file value.
func (Encode) FileKey() string {
	return FileKey
}

// ParseSink provided via form.
// This function should return extensions, sinkbuilder
func (Encode) ParseSink(data url.Values) (fn file.BuildSinkFunc, ext string, err error) {
	ext = strings.ToLower(data.Get("format"))
	switch ext {
	case file.Wav.DefaultExtension:
		fn, err = parseWavSink(data)
	case file.Mp3.DefaultExtension:
		fn, err = parseMp3Sink(data)
	default:
		err = fmt.Errorf("Unsupported format: %v", ext)
	}
	if err != nil {
		ext = ""
	}
	return
}

func parseWavSink(data url.Values) (file.BuildSinkFunc, error) {
	// try to get bit depth
	bitDepth, err := parseIntValue(data, "wav-bit-depth", "bit depth")
	if err != nil {
		return nil, err
	}
	return file.Wav.BuildSink(bitDepth)
}

func parseMp3Sink(data url.Values) (file.BuildSinkFunc, error) {
	// try to get channel mode
	channelMode, err := parseIntValue(data, "mp3-channel-mode", "channel mode")
	if err != nil {
		return nil, err
	}

	// try to get bit rate mode
	bitRateMode := data.Get("mp3-bit-rate-mode")
	if bitRateMode == "" {
		return nil, fmt.Errorf("Please provide bit rate mode")
	}

	var bitRate int
	switch bitRateMode {
	case file.Mp3.VBR:
		// try to get vbr quality
		bitRate, err = parseIntValue(data, "mp3-vbr-quality", "vbr quality")
		if err != nil {
			return nil, err
		}
	case file.Mp3.CBR:
		// try to get bitrate
		bitRate, err = parseIntValue(data, "mp3-bit-rate", "bit rate")
		if err != nil {
			return nil, err
		}
	case file.Mp3.ABR:
		// try to get bitrate
		bitRate, err = parseIntValue(data, "mp3-bit-rate", "bit rate")
		if err != nil {
			return nil, err
		}
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

	return file.Mp3.BuildSink(bitRateMode, bitRate, channelMode, useQuality, quality)
}

// parseIntValue parses value of key provided in the html form.
// Returns error if value is not provided or cannot be parsed as int.
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

// parseBoolValue parses value of key provided in the html form.
// Returns false if value is not provided. Returns error when cannot be parsed as bool.
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
        const fileId = 'input-file';
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
            document.getElementById('input-file').addEventListener('change', onInputFileChange);
            document.getElementById('output-format').addEventListener('change', onOutputFormatChange);
            document.getElementById('submit-button').addEventListener('click', onSubmitClick);
            // mp3 handlers
            document.getElementById('mp3-bit-rate-mode').addEventListener('click', onMp3BitRateModeChange);
            document.getElementById('mp3-use-quality').addEventListener('click', onMp3UseQUalityChange);
        });
        function onInputFileChange(){
            var fileName = getFileName(getFile());
            document.getElementById('input-file-label').innerHTML = fileName;
            var ext = getFileExtension(fileName);
            if (accept.indexOf(ext) < 0) {
                alert('Only files with following extensions are allowed: {{.Accept}}')
                return;
            }
            displayClass('input-file-label', 'inline');
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
            <input id="input-file" type="file" name="input-file" accept="{{.Accept}}"/>
            <label id="input-file-label" for="input-file">select file</label>
        </div>
        <div class="outputs">
            <div id="output-format-block" class="option">
                format
                <select id="output-format" name="format">
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
                    {{range $key, $value := .Wav.BitDepths}}
                        <option value="{{ printf "%d" $key }}">{{ $key }}</option>
                    {{end}}
                </select>
            </div>
            <div id="mp3-options" class="output-options">
                channel mode
                <select name="mp3-channel-mode" class="option">
                    <option hidden disabled selected value>select</option>
                    {{range $key, $value := .Mp3.ChannelModes}}
                        <option value="{{ printf "%d" $key }}">{{ $key }}</option>
                    {{end}}
                </select>
                bit rate mode
                <select id="mp3-bit-rate-mode" class="option" name="mp3-bit-rate-mode">
                    <option hidden disabled selected value>select</option>
                    <option id="{{ .Mp3.VBR  }}" value="{{ .Mp3.VBR }}">{{ .Mp3.VBR }}</option>
                    <option id="{{ .Mp3.CBR  }}" value="{{ .Mp3.CBR }}">{{ .Mp3.CBR }}</option>
                    <option id="{{ .Mp3.ABR  }}" value="{{ .Mp3.ABR }}">{{ .Mp3.ABR }}</option>
                </select>
                <div class="mp3-bit-rate-mode-options mp3-{{ .Mp3.ABR }}-options mp3-{{ .Mp3.CBR }}-options">
                    bit rate [8-320]
                    <input type="text" class="option" name="mp3-bit-rate" maxlength="3" size="3">
                </div>
                <div class="mp3-bit-rate-mode-options mp3-{{ .Mp3.VBR }}-options">
                    vbr quality [0-9]
                    <input type="text" class="option" name="mp3-vbr-quality" maxlength="1" size="3">
                </div>
                <div class="mp3-quality">
                    <input type="checkbox" id="mp3-use-quality" name="mp3-use-quality" value="true">quality
                    <div id="mp3-quality-value" class="mp3-quality" style="visibility:hidden">
                        [0-9]
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
            made with <a href="https://github.com/pipelined/pipe" target="_blank">pipe</a>
            </div>
        </div>
    </div>
</body>
</html>
`
