(function() {
  'use strict';

  angular.module('siteApp')
    .controller('LBNewCtrl', function ($scope, session, $http, $rootScope, $location) {
      session.then(function() {

        $scope.creating = false;

        // because i'm tired of typing this in
        $scope.lbForm = {
          name: 'fe1',
          region: 'tor1',
          sshKeys: '104064',
          enableLogging: true,
          loggingHost: 'logs3.papertrailapp.com',
          loggingPort: 47641,
          loggingSSL: true
        };


        $scope.submitDisabled = function() {
          if ($scope.creating === true) {
            return true;
          }
        };

        $scope.lbForm.submit = function() {
          $scope.errMessage = undefined;
          $scope.creating = true;

          var bootstrapConfig = {
            'digitalocean_token': $rootScope.UserInfo['access_token'], // jshint ignore:line
            name: $scope.lbForm.name,
            region: $scope.lbForm.region,
            'ssh_keys': $scope.lbForm.sshKeys.split(',')
          };

          if ($scope.lbForm.enableLogging) {
            bootstrapConfig['remote_syslog'] = { // jshint ignore:line
              host: $scope.lbForm.loggingHost,
              port: Number($scope.lbForm.loggingPort),
              'enable_ssl': $scope.lbForm.loggingSSL
            };
          }

          $http.post('/api/lb', bootstrapConfig, {}).
            then(function(res) {
              $scope.creating = false;
              console.log(res.load_balancer.id); //jshint ignore:line
              var path = '/lb/' + res['load_balancer'].id; // jshint ignore:line
              $location.path(path);
            }).
            catch(function(e) {
              $scope.errMessage = e.data.error;
              console.log(e);
              $scope.creating = false;
            });
        };
      });

    });
})();
