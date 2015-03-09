/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/**
 * @fileoverview Angularjs controllers for TODO Tracker HTML files.
 */
var todoTrackerApp=angular.module("todoTrackerApp", []);
todoTrackerApp.controller("listRepos", function($scope,$http) {
  $http.get(window.location.protocol + "//" + window.location.host + "/repos")
    .success(function(response) {$scope.repositories = processRepoListResponse(response);});

  function processRepoListResponse(response) {
    var repos = []
    for (var i in response) {
      var repoRaw = response[i]
      repos.push(new Repo(repoRaw.Path, repoRaw.RepoId))
    }
    return repos;
  }

  function Repo(path, id) {
    this.path = path;
    this.id = id;
  }
});

todoTrackerApp.controller("listBranches", function($scope,$http,$location) {
  var repo = $location.search()['repo'];
  $http.get(window.location.protocol + "//" + window.location.host + "/aliases?repo=" + repo)
    .success(function(response) {$scope.remotes = processBranchListResponse(response);});

  function processBranchListResponse(response) {
    var remotesRaw = {};

    for (var i in response) {
      var oneBranchRaw = response[i];
      console.log("branch = " + oneBranchRaw.Branch);
      var result = parseBranchName(oneBranchRaw.Branch);
      // TODO: add lastModified and lastModifiedBy fields
      var branch = new Branch(result[1], oneBranchRaw.Revision, "", "");
      if (!(result[0] in remotesRaw)) {
	remotesRaw[result[0]] = [];
      }
      remotesRaw[result[0]].push(branch);
    }

    var remotes = [];
    for (var r in remotesRaw) {
      var remote = new Remote(r);
      remote.branches = remotesRaw[r];
      remotes.push(remote);
    }

    function Remote(name) {
      this.name = name;
      this.branches = [];
    }

    function Branch(branch, revision, lastModified, lastModifiedBy) {
      this.repo = repo;
      this.branch = branch;
      this.revision = revision;
      this.lastModified = lastModified;
      this.lastModifiedBy = lastModifiedBy;
    }

    function parseBranchName(branchName) {
      var result = branchName.split("/");
      if (result.length >= 3 && result[0] == "remotes") {
        return [result[1], result.slice(2).join('/')];
      } else {
        return ["", result.join('/')];
      }
    }

    console.log("final remotes = " + JSON.stringify(remotes));
    return remotes;
  }
});

todoTrackerApp.controller("listTodos", function($scope,$http,$location) {
  var repo = $location.search()['repo'];
  var revision = $location.search()['revision'];
  $http.get(window.location.protocol + "//" + window.location.host +
      "/revision?repo="+ repo + "&revision=" + revision)
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
      this.repo = repo;
      this.revision = revision;
      this.fileName = fileName;
      this.lineNumber = lineNumber;
      this.content = content;
    }

    return revisionAndTodos;
  }
});

todoTrackerApp.controller("listTodosPaths", function($scope,$http,$location) {
  var repo = $location.search()['repo'];
  var revision = $location.search()['revision'];
  $http.get(window.location.protocol + "//" + window.location.host +
      "/revision?repo="+ repo + "&revision=" + revision)
    .success(function(response) {$scope.filenames = processTodoListPathsResponse(response);});

  function processTodoListPathsResponse(response) {
    var todosObj = response;
    var todosMap = {};

    for (var i = 0; i < todosObj.length; i++) {
      var oneTodoRaw = todosObj[i];
      var fileNameKey = oneTodoRaw.FileName;
      if (!(fileNameKey in todosMap)) {
        todosMap[fileNameKey] = [];
      }
      var todo = new Todo(oneTodoRaw.Revision, oneTodoRaw.FileName,
          oneTodoRaw.LineNumber, oneTodoRaw.Contents);
      todosMap[fileNameKey].push(todo);
    }

    var filenamesAndTodos = [];
    for (var filename in todosMap) {
      var filenameObj = new FileName(filename);
      filenameObj.todos = todosMap[filename];
      filenamesAndTodos.push(filenameObj);
    }


    function FileName(fileName) {
      this.fileName = fileName;
      this.todos = [];
    }


    function Todo(revision, fileName, lineNumber, content) {
      this.repo = repo;
      this.revision = revision;
      this.fileName = fileName;
      this.lineNumber = lineNumber;
      this.content = content;
    }

    return filenamesAndTodos;
  }
});

function getRevisionLink(repo, revision) {
  // the # sign in the URL is to make Angularjs to recoginize QS params in
  // $location.search(). It is a workaround for a bug in Angularjs.
  return window.location.protocol + "//" + window.location.host +
      "/ui/list_todos_paths.html#?repo=" + repo + "&revision=" + revision;
}

todoTrackerApp.controller("todoDetails", function($scope,$http,$location) {
  var repo = $location.search()['repo'];
  var revision = $location.search()['revision'];
  var fileName = $location.search()['fn'];
  var lineNumber = $location.search()['ln'];
  // TODO: Pass in the number of lines above and below the TODO to display
  // This needs the JSON file to provide the informaiton.
  $http.get(window.location.protocol + "//" + window.location.host +
      "/todo?repo=" + repo + "&revision=" + revision +
      "&fileName=" + fileName + "&lineNumber=" + lineNumber)
    .success(function(response) {$scope.todoDetails = processTodoDetailsResponse(response);});

  function processTodoDetailsResponse(response) {
    var detailsObj = response;
    var todoDetails  = [];

    todoDetails.push(new TodoDetail("Revision", detailsObj.Id.Revision, true,
          getRevisionLink(repo, detailsObj.Id.Revision)));
    todoDetails.push(new TodoDetail("File Name", detailsObj.Id.FileName, true,
          getFileInRepoLink(detailsObj.Id.Revision, detailsObj.Id.FileName)));
    todoDetails.push(new TodoDetail("Line Number", detailsObj.Id.LineNumber, true,
          getCodeLineInRepoLink(detailsObj.Id.Revision, detailsObj.Id.FileName, detailsObj.Id.LineNumber)));
    todoDetails.push(new TodoDetail("Author",
          detailsObj.RevisionMetadata.AuthorName + " (" +
          detailsObj.RevisionMetadata.AuthorEmail + ")",
          false, ""));
    todoDetails.push(new TodoDetail("Timestamp",
          timestampPretty(detailsObj.RevisionMetadata.Timestamp) + " (" +
          detailsObj.RevisionMetadata.Timestamp + ")",
          false, ""));
    todoDetails.push(new TodoDetail("Subject", detailsObj.RevisionMetadata.Subject, false, ""));
    // TODO: Display this with syntax highlighting and the TODO line highlighted.
    todoDetails.push(new TodoDetail("Context", detailsObj.Context, false, "", true));

    function TodoDetail(key, value, hasLink, link, htmlPre) {
      this.key = key;
      this.value = value;
      this.hasLink = hasLink;
      this.link = link;
      // Whether to use <pre></pre> on this detail field.
      this.htmlPre = htmlPre == null ? false : htmlPre;
    }

    function getFileInRepoLink(revision, fileName) {
      return window.location.protocol + "//" + window.location.host +
	  "/browse?repo=" + repo + "&revision=" + revision + "&fileName=" + fileName;
    }

    function getCodeLineInRepoLink(revision, fileName, lineNumber) {
      return getFileInRepoLink(revision, fileName) + "&lineNumber=" + lineNumber;
    }

    function timestampPretty(timestamp) {
      var date = new Date(timestamp * 1000);
      return date.toString();
    }

    return todoDetails;
  }
});

todoTrackerApp.controller("todoStatus", function($scope,$http,$location) {
  var repo = $location.search()['repo'];
  var revision = $location.search()['revision'];
  var fileName = $location.search()['fn'];
  var lineNumber = $location.search()['ln'];
  $http.get(window.location.protocol + "//" + window.location.host +
      "/todoStatus?repo=" + repo + "&revision=" + revision +
      "&fileName=" + fileName + "&lineNumber=" + lineNumber)
    .success(function(response) {$scope.todoStatus = processTodoStatusResponse(response);});

  function processTodoStatusResponse(response) {
    var statusObj = response;
    var todoStatus = {present: [], removed: [], missing: []};

    for (var i in statusObj.BranchesPresent) {
      var oneBranchRaw = statusObj.BranchesPresent[i];
      var branch = oneBranchRaw.Branch;
      var revision = oneBranchRaw.Revision;
      var link = getRevisionLink(repo, revision);
      todoStatus.present.push(new BranchDetail(branch, link));
    }

    for (var i in statusObj.BranchesRemoved) {
      var oneBranchRaw = statusObj.BranchesRemoved[i];
      var branch = oneBranchRaw.Branch;
      var revision = oneBranchRaw.Revision;
      var link = getRevisionLink(repo, revision);
      todoStatus.removed.push(new BranchDetail(branch, link));
    }

    for (var i in statusObj.BranchesMissing) {
      var oneBranchRaw = statusObj.BranchesMissing[i];
      var branch = oneBranchRaw.Branch;
      var revision = oneBranchRaw.Revision;
      var link = getRevisionLink(repo, revision);
      todoStatus.missing.push(new BranchDetail(branch, link));
    }

    function BranchDetail(name, link) {
      this.name = name;
      this.link = link;
    }

    return todoStatus;
  }
});
