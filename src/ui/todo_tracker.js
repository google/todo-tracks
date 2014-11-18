/**
 * @fileoverview Description of this file.
 */
var todoTrackerApp = angular.module("todoTrackerApp", []);
todoTrackerApp.controller("listBranches", function($scope) {
  $scope.repositories = [
    {repo_name:'local',
      branches:[
        {branch_name:'master',info:'master_info'},
        {branch_name:'weizheng-dev',info:'weizheng-dev_info'},
        {branch_name:'ojarjur-dev',info:'ojarjur-dev_info'}]},
    {repo_name:'remote',
      branches:[
        {branch_name:'master_on_remote',info:'master_on_remote_info'},
        {branch_name:'weizheng-dev_on_remote',info:'weizheng-dev_on_remote_info'},
        {branch_name:'ojarjur-dev_on_remote',info:'ojarjur-dev_on_remote_info'}]}];
});
