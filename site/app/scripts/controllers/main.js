(function() {
  'use strict';

  angular.module('siteApp')
    .controller('MainCtrl', function ($scope, $cookies, $window, session, $http) {

      var sessionVar = $cookies.get('_dolb_session');
      if (!sessionVar) {
        $window.location.href = '/auth/digitalocean';
        return;
      } 

      session.then(function() {
        $http.get('/api/lb')
          .success(function(res) {
            $scope.lbs=res;
            console.log($scope.lbs);
          })
          .error(function() {
            $scope.lbs={'error': 'could not retrieve load balancers'};
          });
      }, function() {
        console.log('looks like login failed');
      });
    });

})();
