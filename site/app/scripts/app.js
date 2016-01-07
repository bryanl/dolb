'use strict';

/**
 * @ngdoc overview
 * @name siteApp
 * @description
 * # siteApp
 *
 * Main module of the application.
 */


var app = angular.
  module('siteApp', [
      'ngAnimate',
      'ngCookies',
      'ngResource',
      'ngRoute',
      'ngSanitize',
      'ngTouch'
  ]);


app.config(function ($routeProvider, $httpProvider) {
  $routeProvider
    .when('/', {
      templateUrl: 'views/main.html',
      controller: 'MainCtrl'
    })
  .when('/about', {
    templateUrl: 'views/about.html',
    controller: 'AboutCtrl'
  })
  .when('/lb/new', {
    templateUrl: 'views/lb_new.html',
    controller: 'LBNewCtrl'
  })
  .when('/lb/:lbid', {
    templateUrl: 'views/lb_show.html',
    controller: 'LBShowCtrl'
  })
  .otherwise({
    redirectTo: '/'
  });

  $httpProvider.useApplyAsync(true);
  $httpProvider.interceptors.push('httpHandler');
});

app.factory('session', function($http, $q, $rootScope, $log) {
  var defer = $q.defer();

  $http.get('/api/user')
    .then(function(res) {
      $rootScope.UserInfo = res.data;
      $log.debug($rootScope.UserInfo);
      defer.resolve('done');
    }, function(res) {
      $log.debug(JSON.stringify(res));
      defer.reject();
    });

  return defer.promise;
});

app.factory('httpHandler', ['$q', '$log', function($q, $log) {
  var interceptor = {
    'responseError': function(response) {
      switch (response.status) {
        case 0:
          //Do something when we don't get a response back
          break;
        case 401:
          //Do something when we get an authorization error
          break;
        case 400:
          //Do something for other errors
          break;
        case 500:
          //Do something when we get a server error
          break;
        default:
          //Do something with other error codes
          break;
      }
      return $q.reject(response);
    }
  };

  return interceptor;
}]);
