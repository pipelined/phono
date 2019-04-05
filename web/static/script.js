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