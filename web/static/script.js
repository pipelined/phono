$(document).ready(function() {
    document.getElementById("convert").reset();
    $('#input-file').change(function(){
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
            $('#mp3-quality').show().css("display", "inline");
        } else {
            $('#mp3-quality').hide();
        }
    })

    $("#submit").click(function(e){
        var filePath = $('#input-file').val();
        var fileName = filePath.substr(filePath.lastIndexOf('\\') + 1);
        var ext = fileName.split('.')[1];
        $('#convert').attr('action', ext).submit();
    });
});
