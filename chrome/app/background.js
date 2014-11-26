(function () {
  'use strict';

  chrome.tabs.onUpdated.addListener(function(tabId, changeInfo, tab) {
    $.ajax({
      url: "http://localhost:45566/keys",
      type: "POST",
      data: JSON.stringify({url: tab.url}),
      contentType: "application/json; charset=utf-8"
    })
    .done(function(data) {
      chrome.pageAction.show(tabId);
    });
  });

  chrome.contextMenus.create({
    title: 'Save Page Fields',
    contexts: ['all'],
    onclick: function(info, tab) {
      chrome.tabs.sendMessage(tab.id, {type: "GET_FORM"}, function(fields) {
        var form = {
          tabId: tab.id,
          url: tab.url,
          fields: fields
        };
        chrome.windows.create({url: 'newform.html', type: 'popup', width: 400, height: 300}, function(){
          chrome.runtime.sendMessage(form);
        });
      });
    }
  });
}());
