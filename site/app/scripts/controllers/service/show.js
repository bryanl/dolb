(function() {
  'use strict';

  angular.module('siteApp')
    .controller('ServiceShowCtrl', ['$scope', '$log', '$http', '$stateParams', '$state',
        function($scope, $log, $http, $stateParams, $state) {
          $log.debug('stateParams: ' + JSON.stringify($stateParams));

          $scope.lb = $scope.$parent.lb;
          $log.debug($scope.$parent.lb);
        }]);
})();

