'use strict';

var app = require('angular').module('oyster', [
  require('./chrome.js')
]);

app.factory("FormRepo", FormRepo);

function FormRepo($q, Runtime) {
  function list() {
    return $q(function(resolve, reject) {
      Runtime.sendNativeMessage('com.github.proglottis.oyster', {
        type: "LIST"
      }).then(function(response) {
        if (response.type === "ERROR") {
          return reject(response.data);
        }
        resolve(response.data);
      }, reject);
    });
  }

  function search(q) {
    return $q(function(resolve, reject) {
      Runtime.sendNativeMessage('com.github.proglottis.oyster', {
        type: "SEARCH",
        data: {
          query: q
        }
      }).then(function(response) {
        if (response.type === "ERROR") {
          return reject(response.data);
        }
        resolve(response.data);
      }, reject);
    });
  }

  function get(key, password) {
    return $q(function(resolve, reject) {
      Runtime.sendNativeMessage('com.github.proglottis.oyster', {
        type: "GET",
        data: {
          key: key,
          passphrase: password
        }
      }).then(function(response) {
        if (response.type === "ERROR") {
          return reject(response.data);
        }
        resolve(response.data);
      }, reject);
    });
  }

  function put(form) {
    return Runtime.sendNativeMessage('com.github.proglottis.oyster', {
      type: "PUT",
      data: form
    });
  }

  return {list: list, search: search, get: get, put: put};
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
    FormRepo.search(tab.url).then(function(forms) {
      $scope.forms = forms;
      if($scope.forms.length < 1) {
        $scope.message = "No saved forms for this page";
      }
    }, function(err) {
      $scope.message = err;
    });
  });

  $scope.select = function(form) {
    $scope.selectedForm = form;
  };

  $scope.unselect = function() {
    $scope.selectedForm = null;
  };

  $scope.unlock = function() {
    FormRepo.get($scope.selectedForm.key, $scope.password).then(function(forms) {
      Tabs.sendMessage($scope.tabId, {
        type: "SET_FORM",
        data: forms
      });
      $scope.close();
    }, function(err) {
      $scope.message = err;
    });
  };

  $scope.close = function() {
    $window.close();
  };
}

app.controller("FormListCtrl", FormListCtrl);

function FormListCtrl($scope, FormRepo, $log) {
  $scope.password = "";

  FormRepo.list().then(function(forms) {
    $scope.forms = forms;

    if($scope.forms.length < 1) {
      $scope.message = "No saved forms for this page";
    }
  }, function(err) {
    $scope.message = err;
  });

  $scope.edit = function(form) {
    $scope.selectedForm = form;
    $scope.unlocked = false;
  };

  $scope.update = function() {
    FormRepo.put($scope.selectedForm);
    $scope.cancel();
  };

  $scope.cancel = function() {
    $scope.selectedForm = null;
    $scope.unlocked = true;
  };

  $scope.addField = function() {
    if ($scope.selectedForm) {
      $scope.selectedForm.fields.push({});
    }
  };

  $scope.removeField = function(index) {
    if ($scope.selectedForm) {
      $scope.selectedForm.fields.splice(index, 1);
    }
  };

  $scope.unlock = function() {
    FormRepo.get($scope.selectedForm.key, $scope.password).then(function(form) {
      if (form) {
        $scope.selectedForm = form;

        $scope.unlocked = true;
      } else {
        $scope.unlocked = false;
      }
    }, function(err) {
      $scope.message = err;
    });

    $scope.password = "";
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
