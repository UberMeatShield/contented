import { Observable, from as observableFrom } from 'rxjs';
import { Container } from './../../contented/container';
import { Content } from './../../contented/content';
import { ApiDef } from './../../contented/api_def';
import * as _ from 'lodash';

import screensResult from './screens.json';
import splashResult from './splash.json';
import tagsResult from './tags.json';
import videoContentResult from './video_content.json';
import fullResult from './full.json';
import taskRequestsResult from './task_requests.json';
import contentResult from './content.json';
import containersResult from './containers.json';
import searchResult from './search.json';
import videoViewResult from './video_view.json';

declare var require: any;
class MockLoader {
  public timeoutSpan = 100;
  public constructor() {}

  public getPreview() {
    return _.clone(containersResult);
  }

  public getSearch() {
    return _.clone(searchResult);
  }

  public getVideos() {
    return _.clone(videoViewResult);
  }

  // TODO: Get some generated data (for pagination tests)
  public getContainers(total: number = 10) {
    return _.clone(containersResult);
  }

  public getScreens() {
    return _.clone(screensResult);
  }

  public splash() {
    return _.clone(splashResult);
  }

  public tags() {
    return _.clone(tagsResult);
  }

  public videoContent() {
    return _.clone(videoContentResult);
  }

  public getFullContainer() {
    return _.clone(fullResult);
  }

  public taskRequest(taskId: string) {
    let tasks = _.clone(taskRequestsResult);
    let task = tasks[0];
    task.id = taskId;
    return task;
  }

  public taskRequests() {
    return _.clone(taskRequestsResult);
  }

  public getContent(container_id = null, total = null) {
    let res = _.clone(contentResult);
    if (container_id) {
      _.each(res.results, content => {
        content.id = content.id + container_id;
        content.container_id = container_id;
      });
    }
    // TODO: Create fake content / id info if given a count
    return {
      results: res.results.slice(0, total),
      total: total,
    };
  }

  public getContentArr(container_id = null, total = null) {
    let cRes = this.getContent(container_id, total);
    return _.map(cRes.results, r => new Content(r));
  }

  public getMockDir(count: number, itemPrefix: string = 'item-', offset: number = 0, total = 20) {
    let containerId = 'test';
    let contents = _.map(_.range(0, count), idx => {
      let id = idx + offset;
      return { src: itemPrefix + id, id: id, container_id: containerId };
    });

    let fakeDirResponse = {
      total: total,
      path: `narp${containerId}/`,
      name: containerId, // Generate a UUID?
      id: containerId,
      contents: contents, // Note the API does not currently return contents
    };
    return fakeDirResponse;
  }

  public getContentsResponse(count: number, itemPrefix: string = 'item-', offset: number = 0, total = 20) {
    let cnt = this.getMockDir(count, itemPrefix, offset, total);
    return {
      total: total,
      results: cnt.contents,
    };
  }

  public handleCmpDefaultLoad(httpMock, fixture = null) {
    let containers = this.handleContainerLoad(httpMock);
    if (fixture) {
      fixture.detectChanges();
      this.handleContainerContentLoad(httpMock, containers);
    }
  }

  public handleContainerLoad(httpMock): Array<Container> {
    let cntRes = this.getPreview();
    let containersReq = httpMock.expectOne(req => req.url === ApiDef.contented.containers);
    containersReq.flush(cntRes);
    return _.map(cntRes.results, res => new Container(res));
  }

  public handleContainerContentLoad(httpMock, cnts: Array<Container>, count = 2) {
    _.each(cnts, cnt => {
      let url = ApiDef.contented.containerContent.replace('{cId}', cnt.id);
      let reqs = httpMock.match(r => r.url.includes(url));
      _.each(reqs, req => {
        let res = this.getContent(cnt.name, cnt.count);
        req.flush(res);
      });
    });
  }

  public getImg() {
    let res = this.getContent('10', 1);
    let actualContent = res.results[0];
    return new Content(actualContent);
  }

  public getVideo(): Content {
    return new Content(_.clone(videoContentResult));
  }
}
export let MockData = new MockLoader();
