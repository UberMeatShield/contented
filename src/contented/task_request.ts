/*
 * Class for representing the tasks in the background task queue.
 */
import * as _ from 'lodash';
import { z } from 'zod';

export const TASK_STATES = {
  NEW: 'new',
  PENDING: 'pending',
  IN_PROGRESS: 'in_progress',
  CANCELED: 'canceled',
  ERROR: 'error',
  DONE: 'done',
} as const;

export enum TaskOperation {
  ENCODING = 'video_encoding',
  SCREENS = 'screen_capture',
  WEBP = 'webp_from_screens',
  TAGGING = 'tag_content',
  DUPES = 'detect_duplicates',
}

export const TaskOperationEnum = z.enum([TaskOperation.ENCODING, ...Object.values(TaskOperation)]);

export const TaskStatesEnum = z.enum([TASK_STATES.NEW, ...Object.values(TASK_STATES)]);
export type TaskState = z.infer<typeof TaskStatesEnum>;

export const COMPLETE_TASKS: Array<TaskState> = [TASK_STATES.CANCELED, TASK_STATES.ERROR, TASK_STATES.DONE];

export const TaskRequestSchema = z.object({
  id: z.number(),
  content_id: z.number(),
  created_at: z.coerce.date().optional(),
  updated_at: z.coerce.date().optional(),
  started_at: z.coerce.date().optional(),
  status: TaskStatesEnum.default(TASK_STATES.NEW),
  operation: TaskOperationEnum,
  number_of_screens: z.number().optional(),
  start_time_seconds: z.number().optional(),
  codec: z.string().optional(),
  width: z.number().optional(),
  height: z.number().optional(),
  message: z.string().optional(),
  err_msg: z.string().optional(),
});

export type ITaskRequest = z.infer<typeof TaskRequestSchema>;
export class TaskRequest implements ITaskRequest {
  id: number = -1;
  content_id: number = 0;
  created_at: Date | undefined;
  updated_at: Date | undefined;
  started_at: Date | undefined;
  status: TaskState = TASK_STATES.NEW;
  operation: TaskOperation = TaskOperation.ENCODING;
  number_of_screens: number = 0;
  start_time_seconds: number = 0;
  codec: string = '';
  width?: number;
  height?: number;
  message: string = '';
  err_msg: string = '';

  uxLoading = false;

  // For more useful json loading and display of the message
  complexMessage: any;

  constructor(obj: any) {
    this.update(obj);
  }

  update(obj: any) {
    const tr = TaskRequestSchema.parse(obj);
    Object.assign(this, tr);

    if (obj.operation === TaskOperation.DUPES && obj.message) {
      this.complexMessage = JSON.parse(obj.message);
    }
  }

  isComplete(): boolean {
    return COMPLETE_TASKS.includes(this.status);
  }
}
