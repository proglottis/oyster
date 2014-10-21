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
}());
