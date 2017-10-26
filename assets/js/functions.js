Dropzone.autoDiscover = false;
var debug_data, res_data;

$(window).on('load', function () {
    $('.content').addClass('content--loaded');
});

$(document).ready(function () {

    showSERP();
    $('#minio-data').dataTable( {
        "ajax": {
            "url": '/search',
            "dataSrc": ""
        },
       "columns": [
            { "data": "Key" },
            { "data": "Value" }
        ],
       // "sDom": '<"dataTables__top"lfB>rt<"dataTables__bottom"p><"clear">',
        "sDom": 't',
        "bFilter":false,
        "searching":true
    });
    var oTable = $('#minio-data').DataTable();
    $('#search-inp').on('keyup',function(){
        oTable.search($(this).val()).draw();
        showSERP();
    });
    //Dropzone
    if($('.dropzone')[0]) {
        $('.dropzone').dropzone({
            url: '/upload',
            init: function () {

                // Toggle textarea
                this.on('addedfile', function () {
                    $('.text').addClass('text--active');

                    
                });

                // Clear queue
                var dropZone = this;
                $('.text__clear').click(function (e) {
                    e.preventDefault();

                    dropZone.removeAllFiles();
                    $('.text').removeClass('text--active');
                    $('.text__input').val('');

                });
            },
            success: function() {
               
            }
        });
    }

    setInterval(function () {
        //var extracted = $.parseJSON($('.json-text').val());
      //  $('.text__input').val(res_data);
        //autosize.update($('.text__input'));
        $.ajax({
            url: "/results",
            async: false,
            success: function(data){
               // alert(JSON.stringify(data));  
                full_data = JSON.stringify(data.Metadata);
                res_data = JSON.stringify(data.Parsed);
                $('.text__input').val($.parseJSON(res_data));
                oTable.draw();
            }
        });
    }, 600); 
    // Autosize
    if($('.text__input')[0]) {
        //autosize($('.text__input'));
    }

    $('body').on('click', '.text__debug', function (e) {
        e.preventDefault();
        $('.debug').addClass('debug--active');
        clean_json = JSON.parse(full_data);
        $('.debug__input').val(clean_json);
    });

    $('body').on('click', '.debug__close', function (e) {
        e.preventDefault();
        $('.debug').removeClass('debug--active');
    });
});

//function that shows the search results page.
function showSERP() {
    var x = document.getElementById("minio-data");
    if (x.style.display === "none") {
        x.style.display = "inline";
    } else {
        x.style.display = "none";
    }
}