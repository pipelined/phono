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

$(document).ready(function() {
    document.getElementById("convert").reset();

    $(function(){
        $("#upload_link").on('click', function(e){
            e.preventDefault();
            document.getElementById('input-file:hidden').click();
        });
    });

    $('#input-file').change(function(){
        document.getElementById('input-file-label').innerHTML = getFileName('input-file');
        displayClass('input-file-label', true);
        displayId('output-format', "");
    })

    // select output format and show options
    $('.output-formats').click(function(){
        displayClass('output-options', false);
        displayId(this.id+'-options', "");
        displayId('submit', "");
    })

    // select mp3 bit rate mode
    $('#mp3-bit-rate-mode').change(function(){
        displayClass('mp3-bit-rate-mode-options', false);
        var selectedOptions = 'mp3-'+this.options[this.selectedIndex].id+'-options';
        displayClass(selectedOptions, true);
    })

    // use mp3 quality
    $('#mp3-use-quality').change(function(){
        if (this.checked) {
            displayId('mp3-quality-value', "inline");
        } else {
            displayId('mp3-quality-value', "none");
        }
    })

    $("#submit").click(function(e){
        var fileName = getFileName('input-file')
        var ext = fileName.split('.')[1];
        var convert = document.getElementById('convert');
        convert.action = ext;
        convert.submit();
    });
});
