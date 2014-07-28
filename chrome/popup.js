(function () {
  'use strict';

  $(document).on('submit', 'form', function(event){
    event.preventDefault();
    chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
      var tab = tabs[0],
        password = $('#password').val();

      $.post("http://localhost:45566/keys", {url: tab.url, passphrase: password})
      .done(function(data) {
        chrome.tabs.sendMessage(tab.id, {
          type: "SET_PASSWORD",
          text: data.value
        });
        window.close();
      });
    });
  });
}());
