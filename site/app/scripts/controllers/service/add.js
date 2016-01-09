(function() {
  'use strict';

  angular.module('siteApp')
    .controller('ServiceAddCtrl', ['$scope', '$log', function($scope, $log) {
      $scope.serviceForm = {};

      $scope.serviceForm.submit = function() {
        $log.debug('creating service: ' + JSON.stringify($scope.serviceForm));
      };
    }]);
})();

