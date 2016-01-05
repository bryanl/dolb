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
  .config(function ($routeProvider) {
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
