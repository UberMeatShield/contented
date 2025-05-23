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
import { RouterTestingHarness } from '@angular/router/testing';
import { HttpRequest } from '@angular/common/http';
import { HttpTestingController } from '@angular/common/http/testing';

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

  public taskRequest(taskId: number) {
    let tasks = _.clone(taskRequestsResult);
    let task = tasks.results[0];
    task.id = taskId;
    return task;
  }

  public taskRequests() {
    return _.clone(taskRequestsResult);
  }

  public getContent(containerId: number | undefined = undefined, total: number = 2) {
    const contents = _.clone(contentResult);

    let results = contents.results;
    if (containerId) {
      results = results.filter(c => c.container_id === containerId);
    }
    results = results.slice(0, total);
    return {
      results: results,
      total: total,
    };
  }

  public getContentArr(container_id: number | undefined = undefined, total: number | undefined = undefined) {
    let cRes = this.getContent(container_id, total);
    return _.map(cRes.results, r => new Content(r as any));
  }

  public getMockDir(count: number, itemPrefix: string = 'item-', offset: number = 0, total = 20) {
    let containerId = 2;
    let contents = _.map(_.range(0, count), idx => {
      let id = idx + offset;
      return { src: itemPrefix + id, id: id, container_id: containerId };
    });

    let fakeDirResponse = {
      total: total,
      path: `narp${containerId}/`,
      name: containerId.toString(), // Generate a UUID?
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

  public handleCmpDefaultLoad(httpMock: HttpTestingController, fixture: RouterTestingHarness) {
    let containers = this.handleContainerLoad(httpMock);
    if (fixture) {
      fixture.detectChanges();
      this.handleContainerContentLoad(httpMock, containers);
    }
  }

  public handleContainerLoad(httpMock: HttpTestingController): Array<Container> {
    let cntRes = this.getPreview();
    let containersReq = httpMock.expectOne(req => req.url === ApiDef.contented.containers);
    containersReq.flush(cntRes);
    return _.map(cntRes.results, res => new Container(res));
  }

  public handleContainerContentLoad(httpMock: HttpTestingController, cnts: Array<Container>, count = 2) {
    _.each(cnts, cnt => {
      let url = ApiDef.contented.containerContent.replace('{cId}', cnt.id.toString());
      let reqs = httpMock.match((r: HttpRequest<any>) => r.url.includes(url));
      _.each(reqs, req => {
        let res = this.getContent(cnt.id, count);
        req.flush(res);
      });
    });
  }

  public getImg() {
    let img = _.clone(contentResult).results.find(m => m.content_type === 'image/png');
    return new Content(img);
  }

  public getVideo(): Content {
    const contentInfo = _.clone(videoContentResult);
    return new Content(contentInfo);
  }
}
export let MockData = new MockLoader();
