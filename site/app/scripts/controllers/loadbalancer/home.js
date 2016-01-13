(function() {
  'use strict';

  angular.module('siteApp')
    .controller('LBHomeCtrl', 
        ['$scope', '$http', '$stateParams', '$log', '$filter', 'LoadBalancerService', '$state',
          function ($scope, $http, $stateParams, $log, $filter, LoadBalancerService, $state) {
            $log.debug('currentState:' + JSON.stringify($stateParams)); 

            $scope.lb = {};

            $scope.lbID = $stateParams.lbID;

            LoadBalancerService.LoadAll().then(function(data) {
              $scope.lb = $filter('getLbByID')(data.load_balancers, $scope.lbID);
            });


            $scope.lbState = function(state) {
              if ($scope.lb === undefined) {
                return false;
              }
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

