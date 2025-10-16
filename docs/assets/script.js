// Mockzure Documentation Site JavaScript

document.addEventListener('DOMContentLoaded', function() {
    // Mobile navigation toggle
    const navToggle = document.querySelector('.nav-toggle');
    const nav = document.querySelector('.nav');
    
    if (navToggle && nav) {
        navToggle.addEventListener('click', function() {
            nav.classList.toggle('active');
        });
    }
    
    // Smooth scrolling for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });
    
    // Copy code blocks to clipboard
    document.querySelectorAll('pre code').forEach(block => {
        const button = document.createElement('button');
        button.className = 'copy-btn';
        button.textContent = 'Copy';
        button.style.cssText = `
            position: absolute;
            top: 0.5rem;
            right: 0.5rem;
            background: var(--primary-color);
            color: white;
            border: none;
            border-radius: 0.25rem;
            padding: 0.25rem 0.5rem;
            font-size: 0.75rem;
            cursor: pointer;
            opacity: 0;
            transition: opacity 0.2s;
        `;
        
        const pre = block.parentElement;
        pre.style.position = 'relative';
        pre.appendChild(button);
        
        pre.addEventListener('mouseenter', () => {
            button.style.opacity = '1';
        });
        
        pre.addEventListener('mouseleave', () => {
            button.style.opacity = '0';
        });
        
        button.addEventListener('click', async () => {
            try {
                await navigator.clipboard.writeText(block.textContent);
                button.textContent = 'Copied!';
                setTimeout(() => {
                    button.textContent = 'Copy';
                }, 2000);
            } catch (err) {
                console.error('Failed to copy text: ', err);
            }
        });
    });
    
    // Table of contents generation for long pages
    const generateTOC = () => {
        const headings = document.querySelectorAll('h2, h3, h4');
        if (headings.length < 3) return; // Only generate TOC for pages with multiple headings
        
        const toc = document.createElement('div');
        toc.className = 'toc';
        toc.innerHTML = '<h3>Table of Contents</h3><ul></ul>';
        
        const tocList = toc.querySelector('ul');
        
        headings.forEach(heading => {
            const id = heading.textContent.toLowerCase()
                .replace(/[^a-z0-9\s]/g, '')
                .replace(/\s+/g, '-');
            heading.id = id;
            
            const li = document.createElement('li');
            li.className = `toc-${heading.tagName.toLowerCase()}`;
            
            const a = document.createElement('a');
            a.href = `#${id}`;
            a.textContent = heading.textContent;
            
            li.appendChild(a);
            tocList.appendChild(li);
        });
        
        // Insert TOC after the first heading
        const firstHeading = document.querySelector('h1');
        if (firstHeading) {
            firstHeading.parentNode.insertBefore(toc, firstHeading.nextSibling);
        }
    };
    
    // Generate TOC for pages that need it
    const currentPage = window.location.pathname.split('/').pop();
    if (['api-reference.html', 'architecture.html', 'compatibility.html'].includes(currentPage)) {
        generateTOC();
    }
    
    // Search functionality (simple client-side search)
    const createSearchBox = () => {
        const searchBox = document.createElement('div');
        searchBox.className = 'search-box';
        searchBox.innerHTML = `
            <input type="text" placeholder="Search documentation..." id="search-input">
            <div id="search-results" class="search-results"></div>
        `;
        
        const header = document.querySelector('.header-content');
        if (header) {
            header.appendChild(searchBox);
        }
        
        const searchInput = document.getElementById('search-input');
        const searchResults = document.getElementById('search-results');
        
        if (searchInput && searchResults) {
            searchInput.addEventListener('input', function() {
                const query = this.value.toLowerCase();
                if (query.length < 2) {
                    searchResults.innerHTML = '';
                    searchResults.style.display = 'none';
                    return;
                }
                
                const results = searchPage(query);
                if (results.length > 0) {
                    searchResults.innerHTML = results.map(result => 
                        `<a href="#${result.id}">${result.text}</a>`
                    ).join('');
                    searchResults.style.display = 'block';
                } else {
                    searchResults.innerHTML = '<div class="no-results">No results found</div>';
                    searchResults.style.display = 'block';
                }
            });
        }
    };
    
    const searchPage = (query) => {
        const results = [];
        const headings = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
        
        headings.forEach(heading => {
            if (heading.textContent.toLowerCase().includes(query)) {
                results.push({
                    id: heading.id || heading.textContent.toLowerCase().replace(/[^a-z0-9\s]/g, '').replace(/\s+/g, '-'),
                    text: heading.textContent
                });
            }
        });
        
        return results.slice(0, 5); // Limit to 5 results
    };
    
    // Add search box to main pages
    if (['index.html', 'api-reference.html', 'architecture.html'].includes(currentPage)) {
        createSearchBox();
    }
    
    // Dark mode toggle (optional feature)
    const createDarkModeToggle = () => {
        const toggle = document.createElement('button');
        toggle.className = 'dark-mode-toggle';
        toggle.innerHTML = '🌙';
        toggle.title = 'Toggle dark mode';
        
        const header = document.querySelector('.header-content');
        if (header) {
            header.appendChild(toggle);
        }
        
        toggle.addEventListener('click', function() {
            document.body.classList.toggle('dark-mode');
            const isDark = document.body.classList.contains('dark-mode');
            this.innerHTML = isDark ? '☀️' : '🌙';
            localStorage.setItem('darkMode', isDark);
        });
        
        // Load saved preference
        if (localStorage.getItem('darkMode') === 'true') {
            document.body.classList.add('dark-mode');
            toggle.innerHTML = '☀️';
        }
    };
    
    // Uncomment to enable dark mode toggle
    // createDarkModeToggle();
});

// Utility functions
function formatBytes(bytes, decimals = 2) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    });
}

// Export functions for use in other scripts
window.MockzureDocs = {
    formatBytes,
    formatDate
};
