(function() {
  'use strict';

  angular.module('siteApp')
    .controller('ServiceShowCtrl', ['$scope', '$log', '$http', '$stateParams', '$state', 'LoadBalancerService',
        function($scope, $log, $http, $stateParams, $state, LoadBalancerService) {
          $log.debug('stateParams: ' + JSON.stringify($stateParams));

          $scope.service = { config: {}};

          $scope.lb = $scope.$parent.lb;
          $log.debug($scope.$parent.lb);
        }]);
})();

