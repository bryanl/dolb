(function() {
  'use strict';

  angular.module('siteApp')
    .controller('ServiceAddCtrl', ['$scope', '$log', '$http', '$stateParams', function($scope, $log, $http, $stateParams) {
      $log.debug($stateParams);
      $scope.nameRegex = /^[A-Za-z]+[A-Za-z0-9_-]*$/;

      $scope.serviceForm = {port: 80};

      $scope.submit = function() {
        $scope.$broadcast('show-errors-check-validity');

        if ($scope.form.$invalid) { return; }

        $log.debug('creating service: ' + JSON.stringify($scope.serviceForm));
        $http.post('/api/lb/' + $stateParams.lbID + '/services', $scope.serviceForm, {}).
          then(function() {
            $log.debug('service created');
          }).
          catch(function(e) {
            $log.debug(e);
          });
      };
    }]);
})();

