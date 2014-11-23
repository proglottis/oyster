(function () {
  'use strict';
  var app = angular.module('passd', []);

  app.factory("Message", ['$q', function($q) {
    function receive() {
      return $q(function(resolve, reject) {
        chrome.runtime.onMessage.addListener(resolve);
      });
    }

    function sendTab(tabId, message) {
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

    return {receive: receive, sendTab: sendTab};
  }]);

  app.factory("FormRepo", ['$http', function($http) {
    function put(form) {
      return $http.put('http://localhost:45566/keys', form);
    }

    return {put: put}
  }]);

  app.controller("NewFormCtrl", ['$scope', '$window', 'Message', 'FormRepo', function($scope, $window, Message, FormRepo) {
    Message.receive().then(function(form) {
      $scope.tabId = form.tabId;
      $scope.url = form.url;
      $scope.fields = form.fields;
    });

    $scope.addField = function() {
      $scope.fields.push({});
    };

    $scope.removeField = function(index) {
      $scope.fields.splice(index, 1);
    };

    $scope.save = function() {
      var form = {url: $scope.url, fields: $scope.fields};
      Message.sendTab($scope.tabId, {
        type: "SET_FORM",
        data: form
      });
      FormRepo.put(form);
      $scope.close();
    };

    $scope.close = function() {
      $window.close();
    }
  }]);
}());
