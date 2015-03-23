'use strict';

var Uri = require('jsuri');

function urlKey(url) {
    var uri = new Uri(url);
    var key = uri.host();
    if(uri.path() !== "/") {
      key += uri.path();
    }
    return key;
}

function newFormPopup(tabId, url) {
  chrome.tabs.sendMessage(tabId, {type: "GET_FORM"}, function(fields) {
    var form = {
      tabId: tabId,
      key: urlKey(url),
      fields: fields
    };
    chrome.windows.create({url: 'newform.html', type: 'popup', width: 400, height: 500}, function(){
      chrome.runtime.sendMessage(form);
    });
  });
}

chrome.contextMenus.create({
  title: 'Save Page Fields',
  contexts: ['all'],
  onclick: function(info, tab) {
    newFormPopup(tab.id, tab.url);
  }
});
