(function () {
  'use strict';
  var app = angular.module('passd', ['chrome']);

  app.factory("FormRepo", ['$http', function($http) {
    function get(url, password) {
      return $http.post('http://localhost:45566/keys', {url: url}, {
        headers: {
          "Authorization": "Basic " + btoa("passd:" + password)
        }
      });
    }

    function put(form) {
      return $http.put('http://localhost:45566/keys', form);
    }

    return {get: get, put: put};
  }]);

  app.controller("NewFormCtrl", ['$scope', '$window', 'Runtime', 'Tabs', 'FormRepo', function($scope, $window, Runtime, Tabs, FormRepo) {
    Runtime.receive().then(function(form) {
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
  }]);

  app.controller("FormCtrl", ['$scope', '$window', 'Tabs', 'FormRepo', function($scope, $window, Tabs, FormRepo) {

    $scope.fetch = function() {
      Tabs.getCurrentActive().then(function(tab) {
        FormRepo.get(tab.url, $scope.password).then(function(response) {
          Tabs.sendMessage(tab.id, {
            type: "SET_FORM",
            data: response.data
          });
          $scope.close();
        });
      });
    };

    $scope.close = function() {
      $window.close();
    };
  }]);
}());
