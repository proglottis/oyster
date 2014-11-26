(function () {
  'use strict';

  chrome.runtime.onMessage.addListener(function (message, sender, sendResponse) {
    if (!message.type) {
      return;
    }
    switch(message.type) {
    case "SET_FORM":
      message.data.fields.forEach(function(field) {
        $("form [name='"+field.name+"']").val(field.value);
      });
      break;
    case "GET_FORM":
      var fields = $(':input:visible').not('[type=hidden]').serializeArray();
      sendResponse(fields);
      break;
    }
  });

}());
