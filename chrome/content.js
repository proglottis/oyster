(function () {
  'use strict';

  chrome.runtime.onMessage.addListener(function (message, sender, sendResponse) {
    if (message.type && (message.type === "SET_PASSWORD")) {
      for(var key in message.form) {
        if(message.form.hasOwnProperty(key)) {
          $("form [name='"+key+"']").val(message.form[key]);
        }
      }
    }
  });
}());
