/**
 * @fileoverview Description of this file.
 */
var todoTrackerApp=angular.module("todoTrackerApp", []);
todoTrackerApp.controller("listBranches", function($scope,$http) {
  alert("hello, controller");
  $http.get("zz_list_branches_json.html")
    .success(function(response) {$scope.repositories = response;})
    .error(function(data,status,headers,config) {
    alert("failed: data" + data + ", status: " + status};
});
