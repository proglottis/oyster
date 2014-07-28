(function() {
  chrome.runtime.onMessage.addListener(function(message, sender, sendResponse) {
    if (message.type && (message.type == "SET_PASSWORD")) {
      $(':password').first().val(message.text);
    }
  });
})();
