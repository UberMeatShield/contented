import _ from 'lodash';
import { TaskRequest } from './task_request';

export class PageResponse<T>  {
  total: number = 0;
  results: T[] = [];
  initialized: boolean = false;

  constructor(obj: any) {
    this.total = obj.total;
    this.results = _.map(obj.results, (result: any) => new T(result));
    this.initialized = true;
  }
}

export class TaskPageResponse implements IPageResponse<TaskRequest> {
  total: number = 0;
  results: Array<TaskRequest> = [];
  initialized: boolean = false;
  
  constructor(obj: any) {
    this.total = obj.total;
    const tasks = _.map(obj.results, (result: any) => new TaskRequest(result));
    this.results = tasks as Array<TaskRequest>;
    this.initialized = true;
  }
}

// Might have to define _all_ the types...  Annoying.

export type IPageResponse<T> = {
  total: number;
  results: T[];
  initialized: boolean;
};


export function getWindowSize() {
    const width = !(window as any).jasmine ? window.innerWidth : 800;
    const height = !(window as any).jasmine ? window.innerHeight : 800;
    return {width, height};
}
