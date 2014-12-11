'use strict';

function setElementsValueByName(name, value) {
  var elements = document.getElementsByName(name);
  for(var i = 0; i < elements.length; i++) {
    elements[i].value = value;
  }
}

function serializeForms() {
  var fields = [];
  for(var i = 0; i < document.forms.length; i++) {
    var elements = document.forms[i].elements;
    for(var j = 0; j < elements.length; j++) {
      var el = elements[j];
      if(!el.name || el.clientHeight < 1 || el.clientWidth < 1) {
        continue;
      }
      fields.push({name: el.name, value: el.value});
    }
  }
  return fields;
}

chrome.runtime.onMessage.addListener(function (message, sender, sendResponse) {
  switch(message.type) {
  case "SET_FORM":
    message.data.fields.forEach(function(field) {
      setElementsValueByName(field.name, field.value);
    });
    break;
  case "GET_FORM":
    sendResponse(serializeForms());
    break;
  }
});
