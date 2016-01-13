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
      'ngTouch',
      'ui.router'
  ]);

app.config(['$stateProvider', '$urlRouterProvider', function($stateProvider, $urlRouterProvider) {
  $urlRouterProvider.otherwise('/');

  $stateProvider
    .state('home', {
      url: '/',
      templateUrl: 'views/main.html',
      controller: 'MainCtrl',
    })

    .state('lb_new', {
      url: '/lb/new',
      templateUrl: 'views/loadbalancer/new.html',
      controller: 'LBNewCtrl',
    })

    .state('lb', {
      url: '/lb/{lbID}',
      abstract: true,
      templateUrl: 'views/loadbalancer/home.html',
      controller: 'LBHomeCtrl',
    })

    .state('lb.show', {
      url: '',
      templateUrl: 'views/loadbalancer/show.html',
      controller: 'LBShowCtrl',
    })

    .state('lb.add_service', {
      templateUrl: 'views/service/add.html',
      controller: 'ServiceAddCtrl',
    })

    .state('lb.service', {
      url: '/service/{serviceID}',
      templateUrl: 'views/service/show.html',
      controller: 'ServiceShowCtrl',
    })
  ;

}]);

app.config(['$httpProvider', function ($httpProvider) {
  $httpProvider.useApplyAsync(true);
  $httpProvider.interceptors.push('httpHandler');
}]);

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

app.factory('httpHandler', ['$q', function($q) {
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

app.factory('LoadBalancerService', ['$q', '$http', '$log', function($q, $http, $log) {
  var all = [];

  var deferred = $q.defer();
  $http.get('/api/lb').then(function(res) {
    $log.debug(res);
    deferred.resolve(res.data);
    all = res.data;
  }, function(res) {
    $log.debug(res);
    deferred.reject(res);
  });

  return {
    LoadAll: function() {
      return deferred.promise;
    },
    GetAll: function()  {
      return all;
    },
  };

}]);
