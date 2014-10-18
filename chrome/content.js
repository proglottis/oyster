(function () {
  'use strict';

  chrome.runtime.onMessage.addListener(function (message, sender, sendResponse) {
    if (message.type && (message.type === "SET_FORM")) {
      var fields = message.data.fields;
      for(var key in fields) {
        if(!message.data.fields.hasOwnProperty(key)) {
          continue;
        }
        $("form [name='"+key+"']").val(fields[key]);
      }
    }
  });
}());
