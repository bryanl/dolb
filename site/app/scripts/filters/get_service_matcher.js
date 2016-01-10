(function(){
  'use strict';

  angular.module('siteApp')
    .filter('getServiceMatcher', function() {
      return function(service) {
        return service.config.matcher;
      };
    });
})();
