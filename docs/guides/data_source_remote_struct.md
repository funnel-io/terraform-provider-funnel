----
page_title: "Data source `remote_struct` for each data source type"
subcategory: "Data source setup"
---

# Data source `remote_struct` for the most popular data sources

Each source type (e.g., `facebookads` and `adwords`) have a defined way to configure which data you will download from the platforms. Some platforms you only need to provide `remote_id` but for some platforms you will need to provide a JSON encoded string with the fields that are specified below.

Required fields for the `remote_struct` are marked with a "*".

## Data source types configured with "remote_id"

### Facebook Ads (`facebookads`)

Facebook Ad Account ID

### Microsoft Advertising (`bing`)

The ad account ID

### TikTok (`tiktok`)

The advertiser ID

### LinkedIn (`linkedin_api`)

The ad account ID

### Pinterest (`pinterest`)

The advertiser ID

### X Ads (`twitter`)

The ad account ID

### Facebook Pages (`facebookpages`)

The page/account ID

### Snapchat (`snapchat`)

The ad account ID

### Instagram Insights (`instagraminsights`)

The page/account ID

### LinkedIn Organic (`linkedin_organic`)

The organization ID

### Apple Search Ads (`applesearchads_api`)

Apple Search Ads organization ID

### Shopify (`shopify`)

The shop ID

Pattern: `^gid://shopify/[A-Za-z]+/[0-9]+$`

### Reddit (`reddit`)

The ad account ID

### X Organic (`twitterorganic`)

The account ID

### Klaviyo (`klaviyo`)

The Klaviyo account ID

### HubSpot Contacts (`hubspot_contacts`)

The portal ID

### Google My Business (`googlemybusiness`)

The page ID

Pattern: `^accounts/[0-9]+$`

### Spotify Ad Studio (`spotifyads`)

The ad account ID

## Data source types configured with "remote_struct"

### Google Ads (`adwords`)

| Field | Type | Description |
| --- | --- | --- |
| `customerId` | string* | Google Ads Customer ID |
| `mccId` | string | MCC (manager) account ID, if the account is managed |

Remote ID formula: `{{customerId}}`

### Google Analytics (`googleanalytics`)

| Field | Type | Description |
| --- | --- | --- |
| `remote_id` | string* | GA4 property ID |
| `remote_parent_id` | string* | The GA4 account ID that owns the property |

### Google Search Console (`googlesearchconsole`)

| Field | Type | Description |
| --- | --- | --- |
| `siteUrl` | string* | Site URL or domain |

Remote ID formula: `{{sha256_16 siteUrl}}`

### YouTube (`youtube`)

| Field | Type | Description |
| --- | --- | --- |
| `channelId` | string* | YouTube channel ID |

Remote ID formula: `{{sha256_16 channelId}}`

### Amazon Ads (`amazonadvertising`)

| Field | Type | Description |
| --- | --- | --- |
| `profileId` | string* | Advertising API profile ID |
| `accountType` | (vendor \| seller)* | Account type |
| `region` | (eu \| us \| far_east)* | API region for the account |

Remote ID formula: `{{profileId}}`

## Data source types that are not supported

- Google Sheets (`googlesheets`)
