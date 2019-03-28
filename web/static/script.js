$(document).ready(function() {
    $('#wav').click(function () {
        if (this.checked) {
            $('.output-options').hide();
            $('#wav-options').show();
        }
    });

    $('#mp3').click(function () {
        if (this.checked) {
            $('.output-options').hide();
            $('#mp3-options').show();
        }
    });  
});
