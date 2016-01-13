(function() {
  'use strict';

  angular.module('siteApp')
    .filter('getLbByID', function() {
      return function(lbs, id) {
        var lb;
        for (lb of lbs) {
          if (lb.id === id) {
            console.log('returning: ' + lb.id);
            return lb;
          }
        }
      };
    });
})();
