/**
 * @fileoverview Description of this file.
 */
var todoTrackerApp=angular.module("todoTrackerApp", []);
todoTrackerApp.controller("listBranches", function($scope,$http) {
  // $http.get(window.location.protocol + "//" + window.location.host + "/zz_list_branches_json.html")
  $http.get(window.location.protocol + "//" + window.location.host + "/aliases")
    .success(function(response) {$scope.repositories = processBranchListResponse(response);});
});

function processBranchListResponse(response) {
  var branchesStrObj = response;
  var reposRaw = {};

  for (var i = 0; i < branchesStrObj.length; i++) {
    var oneBranchRaw = branchesStrObj[i];
    console.log("branch = " + oneBranchRaw.Branch);
    var result = parseBranchName(oneBranchRaw.Branch);
    // TODO: add lastModified and lastModifiedBy fields
    var branch = new Branch(result[1], oneBranchRaw.Revision, "", "");
    if (!(result[0] in reposRaw)) {
      reposRaw[result[0]] = [];
    }
    reposRaw[result[0]].push(branch);
  }

  var repos = [];
  for (var r in reposRaw) {
    var repo = new Repository(r);
    repo.branches = reposRaw[r];
    repos.push(repo);
  }


  function Repository(repository) {
    this.repository = repository;
    this.branches = [];
  }

  function Branch(branch, revision, lastModified, lastModifiedBy) {
    this.branch = branch;
    this.revision = revision;
    this.lastModified = lastModified;
    this.lastModifiedBy = lastModifiedBy;
  }

  function parseBranchName(branchName) {
    var result = branchName.split("/");
    if (result.length == 3) {
      return [result[0], result[2]];
    } else {
      return ['local', result[0]];
    }
  }

  console.log("final repos = " + JSON.stringify(repos));
  return repos;
}

/*
todoTrackerApp.controller("viewBranch", function($scope,$http) {
  $http.get(window.location.protocol + "//" + window.location.host + "/aliases")
    .success(function(response) {$scope.repositories = processBranchListResponse(response);});
});
*/
