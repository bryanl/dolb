'use strict';

/**
 * @ngdoc overview
 * @name siteApp
 * @description
 * # siteApp
 *
 * Main module of the application.
 */
angular
  .module('siteApp', [
    'ngAnimate',
    'ngCookies',
    'ngResource',
    'ngRoute',
    'ngSanitize',
    'ngTouch'
  ])
  .config(function ($routeProvider, $locationProvider) {
    $routeProvider
      .when('/', {
        templateUrl: 'views/main.html',
        controller: 'MainCtrl'
      })
      .when('/about', {
        templateUrl: 'views/about.html',
        controller: 'AboutCtrl'
      })
      .otherwise({
        redirectTo: '/'
      });

      $locationProvider.html5Mode(true);
  })
  .factory('session', function($http, $q, $rootScope) {
    var defer = $q.defer();

    $http.get('/api/user')
      .success(function(res) {
        console.log(res);
        $rootScope.UserInfo = res;
        defer.resolve('done');
      })
      .error(function() {
        defer.reject();
      });

    return defer.promise;
  });
