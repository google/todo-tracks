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
  var branchesObj = response;
  var reposRaw = {};

  for (var i = 0; i < branchesObj.length; i++) {
    var oneBranchRaw = branchesObj[i];
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

todoTrackerApp.controller("listTodos", function($scope,$http,$location) {
  console.log("location = " + $location + ", search = " + $location.search() + ", revid=" + $location.search()['revid']);
  console.log("location = " + JSON.stringify($location));
  console.log("http = " + JSON.stringify($http));

  $http.get(window.location.protocol + "//" + window.location.host + "/revision?id=" + $location.search()['revid'])
    .success(function(response) {$scope.revisions= processTodoListResponse(response);});

   function processTodoListResponse(response) {
    var todosObj = response;
    var todosMap = {};

    for (var i = 0; i < todosObj.length; i++) {
      var oneTodoRaw = todosObj[i];
      if (!(oneTodoRaw.Revision in todosMap)) {
        todosMap[oneTodoRaw.Revision] = [];
      }
      var todo = new Todo(oneTodoRaw.Revision, oneTodoRaw.FileName,
          oneTodoRaw.LineNumber, oneTodoRaw.Contents);
      todosMap[oneTodoRaw.Revision].push(todo);
    }

    var revisionAndTodos = [];
    for (var revisionId in todosMap) {
      var revision = new Revision(revisionId);
      revision.todos = todosMap[revisionId];
      revisionAndTodos.push(revision);
    }


    function Revision(revision) {
      this.revision = revision;
      this.todos = [];
    }

    function Todo(revision, fileName, lineNumber, content) {
      this.revision = revision;
      this.fileName = fileName;
      this.lineNumber = lineNumber;
      this.content = content;
    }

    return revisionAndTodos;
  }
});
