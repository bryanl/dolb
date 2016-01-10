(function() {
  'use strict';

  angular.module('siteApp')
    .controller('LBShowCtrl', function (session, $scope, $routeParams, $http, $location, $stateParams, $log, $state) {
      session.then(function() {
        $log.debug('currentState:' + JSON.stringify($stateParams)); 

        $scope.lbID = $stateParams.lbID;
        $scope.lb = {};

        $scope.deleting = false;

        $scope.canAddService = true;
        $scope.$watch('lb.state', function(val) {
          $scope.canAddService = (val !== 'up');
        });

        var u = '/api/lb/' + $scope.lbID;
        $http.get(u)
          .success(function(res) {
            $scope.lb = res;
          })
          .error(function(res) {
            console.log('error: ' + JSON.stringify(res));
          });

        $scope.addService = function() {
          $state.go('lb.add_service');
        };

        $scope.deleteDisabled = function() {
          if ($scope.deleting === true) {
            return true;
          }
        };

        $scope.deleteLB = function() {
          $scope.deleting = true;

          var u = '/api/lb/' + $scope.lbID;
          $http.delete(u)
            .success(function() {
              $state.go('home');
            })
            .error(function(data) {
              console.log('lb delete failed: ' + JSON.stringify(data));
            });
        };
      });
    });

})();

