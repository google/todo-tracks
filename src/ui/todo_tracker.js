/**
 * @fileoverview Description of this file.
 */
var todoTrackerApp=angular.module("todoTrackerApp", []);
todoTrackerApp.controller("listBranches", function($scope,$http) {
  $http.get(window.location.protocol + "//" + window.location.host + "/zz_list_branches_json.html")
    .success(function(response) {$scope.repositories = response;});
});
