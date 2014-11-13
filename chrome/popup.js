(function () {
  'use strict';

  $(document).on('submit', 'form', function(event){
    event.preventDefault();
    chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
      var tab = tabs[0],
        password = $('#password').val();

      $.ajax({
        url: "http://localhost:45566/keys",
        type: "POST",
        data: JSON.stringify({url: tab.url}),
        contentType: "application/json; charset=utf-8",
        headers: {
          "Authorization": "Basic " + btoa("passd:" + password)
        }
      })
      .done(function(data) {
        var fields = data.fields;
        data.fields = Object.keys(fields).map(function(key) { return {name: key, value: fields[key]} });
        chrome.tabs.sendMessage(tab.id, {
          type: "SET_FORM",
          data: data
        });
        window.close();
      });
    });
  });
}());
