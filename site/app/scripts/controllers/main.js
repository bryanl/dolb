(function() {
  'use strict';

  angular.module('siteApp')
    .controller('MainCtrl', function ($scope, $cookies, $window, session) {

      var sessionVar = $cookies.get('_dolb_session');
      if (!sessionVar) {
        $window.location.href = '/auth/digitalocean';
        return;
      } 

      session.then(function() {
        $scope.todos = ['Item 1', 'Item 2', 'Item 3', 'Item 4'];
      });
    });

})();
