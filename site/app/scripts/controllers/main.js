(function() {
  'use strict';

  angular.module('siteApp')
    .controller('MainCtrl', function ($scope, $cookies, $window, session, $http, LoadBalancerService) {

      var sessionVar = $cookies.get('_dolb_session');
      if (!sessionVar) {
        $window.location.href = '/auth/digitalocean';
        return;
      } 

      session.then(function() {
        $scope.lbs = [];

        LoadBalancerService.LoadAll().then(function(lbs) {
          $scope.lbs = lbs;
        }, function() {
          console.log('load all failed');
        });
      }, function() {
        console.log('looks like login failed');
      });
    });

})();
