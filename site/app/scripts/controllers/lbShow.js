(function() {
  'use strict';

  angular.module('siteApp')
    .controller('LBShowCtrl', function (session, $scope, $routeParams, $http) {
      session.then(function() {
        $scope.lbID = $routeParams.lbid;
        $scope.lb = {};

        var u = '/api/lb/' + $scope.lbID;
        console.log('fetching: ' + u);
        $http.get(u)
          .success(function(res) {
            console.log(res);
            $scope.lb = res;
          })
          .error(function(res) {
            console.log('error: ' + JSON.stringify(res));
          });
      });
    });

})();

