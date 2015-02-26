'use strict';

var app = require('angular').module('oyster', [
  require('./chrome.js')
]);

app.factory("FormRepo", FormRepo);

function FormRepo($http, $window) {
  var base = 'http://localhost:45566/v1';

  function search(q) {
    return $http.get(base + '/keys', {params: {q: q}});
  }

  function get(key, password) {
    return $http.post(base + '/keys', {key: key}, {
      headers: {
        "Authorization": "Basic " + $window.btoa("oyster:" + password)
      }
    });
  }

  function put(form) {
    return $http.put(base + '/keys', form);
  }

  return {search: search, get: get, put: put};
}

app.controller("NewFormCtrl", NewFormCtrl);

function NewFormCtrl($scope, $window, Runtime, Tabs, FormRepo) {
  // Receive from background context menu handler
  Runtime.receive().then(function(form) {
    $scope.tabId = form.tabId;
    $scope.key = form.key;
    $scope.fields = form.fields;
  });

  $scope.addField = function() {
    $scope.fields.push({});
  };

  $scope.removeField = function(index) {
    $scope.fields.splice(index, 1);
  };

  $scope.save = function() {
    var form = {key: $scope.key, fields: $scope.fields};
    Tabs.sendMessage($scope.tabId, {
      type: "SET_FORM",
      data: form
    });
    FormRepo.put(form);
    $scope.close();
  };

  $scope.close = function() {
    $window.close();
  };
}

app.controller("FormSearchCtrl", FormSearchCtrl);

function FormSearchCtrl($scope, Tabs, FormRepo, $window) {
  $scope.password = "";
  Tabs.getCurrentActive().then(function(tab) {
    $scope.tabId = tab.id;
    FormRepo.search(tab.url).then(function(response) {
      $scope.forms = response.data;
      if($scope.forms.length < 1) {
        $scope.message = "No saved forms for this page";
      }
    }, function() {
      $scope.message = "Cannot connect to Oyster";
    });
  });

  $scope.select = function(form) {
    $scope.selectedForm = form;
  };

  $scope.unselect = function() {
    $scope.selectedForm = null;
  };

  $scope.unlock = function() {
    FormRepo.get($scope.selectedForm.key, $scope.password).then(function(response) {
      Tabs.sendMessage($scope.tabId, {
        type: "SET_FORM",
        data: response.data
      });
      $scope.close();
    }, function(response, statuscode) {
      switch(statuscode) {
      default:
        $scope.message = "Cannot connect to Oyster";
      }
    });
  };

  $scope.close = function() {
    $window.close();
  };
}

app.directive('focus', focus);

function focus($timeout) {
  function link(scope, element, attr) {
    scope.$watch(attr.focus, function(focusVal) {
      $timeout(function() {
        if(focusVal) {
          element[0].focus();
        } else {
          element[0].blur();
        }
      });
    });
  }

  return {
    restrict : 'A',
    link : link
  };
}
