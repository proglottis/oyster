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
      chrome.browserAction.setBadgeText({text: "1", tabId: tabId});
    })
    .fail(function(jqXHR) {
      if (jqXHR.status !== 404) {
        chrome.browserAction.setBadgeText({text: "Fail", tabId: tabId});
      }
    });
  });
}());
