(function() {
  'use strict';

  angular.module('siteApp')
    .controller('LBShowCtrl', function (session, $scope, $routeParams, $http, $location) {
      session.then(function() {
        $scope.lbID = $routeParams.lbid;
        $scope.lb = {};

        var u = '/api/lb/' + $scope.lbID;
        $http.get(u)
          .success(function(res) {
            $scope.lb = res;
          })
          .error(function(res) {
            console.log('error: ' + JSON.stringify(res));
          });

        $scope.deleteLB = function() {
          var u = '/api/lb/' + $scope.lbID;
          $http.delete(u)
            .success(function() {
              $location.path('/');
            })
            .error(function(data) {
              console.log('lb delete failed: ' + JSON.stringify(data));
            });
        };
      });
    });

})();

