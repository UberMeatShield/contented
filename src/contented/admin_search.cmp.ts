import { SearchCmp } from './search.cmp';
import {
    Component,
} from '@angular/core';
import {Content} from './content';


@Component({
    selector: 'admin-search-cmp',
    templateUrl: './search.ng.html'
})
export class AdminSearchCmp extends SearchCmp {

    contentClicked(mc: Content) {
        console.log("Click a content element to open the editor for it.");
        window.open(`/admin_ui/editor_content/${mc.id}`);
        window.focus();
    }
}
