import { Observable, from as observableFrom } from 'rxjs';
import { Container } from './../../contented/container';
import { Content } from './../../contented/content';
import { ApiDef } from './../../contented/api_def';
import * as _ from 'lodash';

declare var require: any;
class MockLoader {
  public timeoutSpan = 100;
  public constructor() {}

  public getPreview() {
    return _.clone(require('./containers.json'));
  }

  public getSearch() {
    return _.clone(require('./search.json'));
  }

  public getVideos() {
    return _.cloneDeep(require('./video_view.json'));
  }

  public getScreens() {
    return _.clone(require('./screens.json'));
  }

  public splash() {
    return _.clone(require('./splash.json'));
  }

  public tags() {
    return _.clone(require('./tags.json'));
  }

  public videoContent() {
    return _.clone(require('./video_content.json'));
  }

  public getFullContainer() {
    return _.clone(require('./full.json'));
  }

  public taskRequest(taskId: string) {
    let tasks = _.clone(require('./task_requests.json'));
    let task = tasks[0];
    task.id = taskId;
    return task;
  }

  public taskRequests() {
    return _.clone(require('./task_requests.json'));
  }

  public getContent(container_id = null, total = null) {
    let res = _.clone(require('./content.json'));
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
}
export let MockData = new MockLoader();
