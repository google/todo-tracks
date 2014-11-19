/**
 * @fileoverview Description of this file.
 */
var todoTrackerApp = angular.module("todoTrackerApp", []);
todoTrackerApp.controller("listBranches", function($scope) {
  $scope.repositories = [
    {repo_name:'local',
      branches:[
        {branch_name:'master',last_modified:'Nov 10 by Wei',link:'xx'},
        {branch_name:'weizheng-dev',last_modified:'Nov 10 by Wei',link:'xx'},
        {branch_name:'ojarjur-dev',last_modified:'Nov 10 by Wei',link:'xx'}]},
    {repo_name:'remote',
      branches:[
        {branch_name:'master_on_remote',last_modified:'Nov 10 by Wei',link:'xx'},
        {branch_name:'weizheng-dev_on_remote',last_modified:'Nov 10 by Wei',link:'xx'},
        {branch_name:'ojarjur-dev_on_remote',last_modified:'Nov 10 by Wei',link:'xx'}]}];
});
