(function() {

  $(document).on('submit', 'form', function(event){
    event.preventDefault();
    chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
      var tab = tabs[0],
        password = $('#password').val();

      $.post("http://localhost:45566/keys", {url: tab.url, passphrase: password})
      .done(function(data) {
        chrome.tabs.executeScript({
          code:
            "(function() {"
          + "var inputs = document.getElementsByTagName('input');"
          + "for (var i=0; i<inputs.length; i++) {"
          + "  if (inputs) {"
          + "    if (inputs[i].type.toLowerCase() === 'password') {"
          + "      inputs[i].value = '" + data.value + "';"
          + "      break;"
          + "    }"
          + "  }"
          + "}"
          + "})();"
        });
        window.close();
      });
    });
  });

})();
