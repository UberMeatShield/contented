import {
    OnInit,
    Component,
    Input,
    Output,
    ViewChild,
    EventEmitter,
} from '@angular/core';
import { Tag, VSCodeChange } from './content';
import {ContentedService} from './contented_service';

import * as _ from 'lodash';
import { GlobalBroadcast } from './global_message';


const editorOptions = {
    theme: 'vs-dark',
    renderLineHighlight: "none",
    //quickSuggestions: false,
    glyphMargin: false,
    lineDecorationsWidth: 0,
    folding: false,
    fixedOverflowWidgets: true,
    acceptSuggestionOnEnter: "on",
    placeholder: 'Search',
    hover: {
      delay: 100,
    },
    roundedSelection: false,
    contextmenu: false,
    cursorStyle: "line-thin",
    occurrencesHighlight: false,
    links: false,
    minimap: { enabled: false },
    // see: https://github.com/microsoft/monaco-editor/issues/1746
    wordBasedSuggestions: false,
    // disable `Find`
    find: {
      addExtraSpaceOnTop: false,
      autoFindInSelection: "never",
      seedSearchStringFromSelection: "never",
    },
    fontSize: 14,
    fontWeight: "normal",
    wordWrap: "off",
    lineNumbers: "off",
    lineNumbersMinChars: 0,
    overviewRulerLanes: 0,
    overviewRulerBorder: false,
    hideCursorInOverviewRuler: true,
    scrollBeyondLastColumn: 0,
    scrollbar: {
      horizontal: "hidden",
      vertical: "hidden",
      // avoid can not scroll page when hover monaco
      alwaysConsumeMouseWheel: false,
    },
    language: 'tagging',
};

@Component({
    selector: 'tags-cmp',
    templateUrl: './tags.ng.html'
})
export class TagsCmp implements OnInit{

    // Route needs to exist
    // Take in the search text route param
    // Debounce the search
    @ViewChild('searchForm', { static: true }) searchControl;

    @Input() editorValue: string = "";
    @Input() tags: Array<Tag>;
    @Input() loadTags = false;
    @Input() editorOptions;

    @Output() tagsChanged = new EventEmitter<VSCodeChange>;

    matchedTags: Array<Tag>;


    public loading: boolean = false;
    public pageSize: number = 1000;
    public total: number = 0;

    constructor(
        public _contentedService: ContentedService,
    ) {
        this.editorOptions = this.editorOptions || editorOptions;
    }

    public ngOnInit() {
        if (this.loadTags) {
            this.search('');
        }
    }

    // TODO: Get the current input token
    // TODO: Suggest input tokens
    // TODO: Provide the ability to select tokens and also remove a token from VSCode
    public search(searchText: string) {
        this._contentedService.getTags().subscribe({
            next: (res: any ) => {
                this.tags = res.results;
            },
            error: (err) => {
                GlobalBroadcast.error('Failed to load tags in tagging component', err.message);
            }
        });
    }

    // Change the event to provide both the value and the parsed tags
    public changedTags(evt: VSCodeChange) {
        console.log("Changed Tags", evt);
        this.tagsChanged.emit(evt);
    }
}