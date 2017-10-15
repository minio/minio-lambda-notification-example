$(function() {
   alert("hey1");
    $.ajax({
        url: 'localhost:3000/results',
        success: function(data) {
            alert("hey");
          $('#term').html(data);
        }
      });
});