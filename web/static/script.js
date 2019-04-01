function getFileName(fileId) {
    var filePath = $(fileId).val();
    return filePath.substr(filePath.lastIndexOf('\\') + 1);
}

$(document).ready(function() {
    document.getElementById("convert").reset();

    $(function(){
        $("#upload_link").on('click', function(e){
            e.preventDefault();
            $("#input-file:hidden").trigger('click');
        });
    });

    $('#input-file').change(function(){
        $('#input-file-label').text(getFileName('#input-file'));
        $('#output-formats').show();
    })

    // select output format and show options
    $('.format').click(function(){
        $('.output-options').hide();
        $('#'+this.id+'-options').show();
        $('#submit').show();
    })

    // select mp3 bit rate mode
    $('#mp3-bit-rate-mode').change(function(){
        $('.mp3-bit-rate-mode-options').hide();
        $('.mp3-'+this.value+'-options').show();
    })

    // use mp3 quality
    $('#mp3-use-quality').change(function(){
        if (this.checked) {
            $('#mp3-quality-value').show().css("display", "inline");
        } else {
            $('#mp3-quality-value').hide();
        }
    })

    $("#submit").click(function(e){
        var fileName = getFileName('#input-file')
        var ext = fileName.split('.')[1];
        $('#convert').attr('action', ext).submit();
    });
});
