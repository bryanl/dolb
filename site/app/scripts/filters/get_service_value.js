(function(){
  'use strict';

  angular.module('siteApp')
    .filter('getServiceValue', function() {
      return function(service) {
        return service.config[service.config.matcher];
      };
    });
})();
