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

todoTrackerApp.controller("todoDetails", function($scope,$http,$location) {
  var revision = $location.search()['revid'];
  var fileName = $location.search()['fn'];
  var lineNumber = $location.search()['ln'];
  // TODO: Pass in the number of lines above and below the TODO to display
  $http.get(window.location.protocol + "//" + window.location.host +
      "/todo?revision=" + revision + "&fileName=" + fileName + "&lineNumber=" + lineNumber)
    .success(function(response) {$scope.todoDetails = processTodoDetailsResponse(response);});

  function processTodoDetailsResponse(response) {
    var detailsObj = response;
    var todoDetails  = [];

    todoDetails.push(new TodoDetail("Revision", detailsObj.Id.Revision, true,
          getRevisionLink(detailsObj.Id.Revision)));
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
      // use <pre></pre> on this detail field.
      this.htmlPre = htmlPre == null ? false : htmlPre; // use <pre></pre> on this detail field.
    }

    function Todo(revision, fileName, lineNumber, content) {
      this.revision = revision;
      this.fileName = fileName;
      this.lineNumber = lineNumber;
      this.content = content;
    }

    function getRevisionLink(revision) {
      return window.location.protocol + "//" + window.location.host + "/ui/list_todos.html#?revid=" + revision;
    }

    function getFileInRepoLink(revision, fileName) {
      return window.location.protocol + "//" + window.location.host + "/browse?revision=" + revision +
          "&fileName=" + fileName;
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
