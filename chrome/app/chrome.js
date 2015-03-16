'use strict';

module.exports = 'chrome';
var c = require('angular').module('chrome', []);

c.factory("Tabs", Tabs);

function Tabs($q) {
  function sendMessage(tabId, message) {
    return $q(function(resolve, reject) {
      chrome.tabs.sendMessage(tabId, message, function(response) {
        if (chrome.runtime.lastError) {
          reject(chrome.runtime.lastError.message);
          return;
        }
        resolve(response);
      });
    });
  }

  function getCurrentActive() {
    return $q(function(resolve, reject) {
      chrome.tabs.query({active: true, currentWindow: true}, function(tabs) {
        if (chrome.runtime.lastError) {
          reject(chrome.runtime.lastError.message);
          return;
        }
        if (tabs.length < 1) {
          reject({message: "No active tab"});
          return;
        }
        resolve(tabs[0]);
      });
    });
  }

  return {sendMessage: sendMessage, getCurrentActive: getCurrentActive};
}

c.factory("Runtime", Runtime);

function Runtime($q) {
  function sendNativeMessage(application, message) {
    return $q(function(resolve, reject) {
      chrome.runtime.sendNativeMessage(application, message, function(response) {
        if (chrome.runtime.lastError) {
          reject(chrome.runtime.lastError.message);
          return;
        }
        resolve(response);
      });
    });
  }

  function receive() {
    return $q(function(resolve, reject) {
      chrome.runtime.onMessage.addListener(function(message, sender, sendResponse) {
        if (chrome.runtime.lastError) {
          reject(chrome.runtime.lastError.message);
          return;
        }
        resolve(message, sender, sendResponse);
      });
    });
  }

  return {sendNativeMessage: sendNativeMessage, receive: receive};
}
