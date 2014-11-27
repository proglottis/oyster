'use strict';

var Uri = require('jsuri');

chrome.runtime.onInstalled.addListener(function() {
  chrome.contextMenus.create({
    title: 'Save Page Fields',
    contexts: ['all'],
    onclick: function(info, tab) {
      chrome.tabs.sendMessage(tab.id, {type: "GET_FORM"}, function(fields) {
        var uri = new Uri(tab.url);
        var form = {
          tabId: tab.id,
          key: uri.host() + uri.path(),
          fields: fields
        };
        chrome.windows.create({url: 'newform.html', type: 'popup', width: 400, height: 300}, function(){
          chrome.runtime.sendMessage(form);
        });
      });
    }
  });
});
