(function() {
  'use strict';

  angular.module('siteApp')
    .controller('LBHomeCtrl', 
        ['$scope', '$http', '$stateParams', '$log',
          function ($scope, $http, $stateParams, $log) {
            $log.debug('currentState:' + JSON.stringify($stateParams)); 

            $scope.lbID = $stateParams.lbID;
            $scope.lb = {};

            var u = '/api/lb/' + $scope.lbID;
            $http.get(u)
              .success(function(res) {
                $scope.lb = res;
              })
            .error(function(res) {
              console.log('error: ' + JSON.stringify(res));
            });

            $scope.lbState = function(state) {
              return $scope.lb.state === state;
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
      }]);
})();

