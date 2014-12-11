'use strict';

var c = require('angular').module('chrome', []);

c.factory("Tabs", Tabs);

function Tabs($q) {
  function sendMessage(tabId, message) {
    return $q(function(resolve, reject) {
      chrome.tabs.sendMessage(tabId, message, function(response) {
        var lastError = chrome.runtime.lastError;
        if (lastError) {
          reject(lastError);
        } else {
          resolve(response);
        }
      });
    });
  }

  function getCurrentActive() {
    return $q(function(resolve, reject) {
      chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
        if(tabs.length > 0) {
          resolve(tabs[0]);
        } else {
          reject();
        }
      });
    });
  }

  return {sendMessage: sendMessage, getCurrentActive: getCurrentActive};
}

c.factory("Runtime", Runtime);

function Runtime($q) {
  function receive() {
    return $q(function(resolve, reject) {
      chrome.runtime.onMessage.addListener(resolve);
    });
  }

  return {receive: receive};
}
